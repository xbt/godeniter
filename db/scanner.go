// Copyright (c) 2026 xbt. All rights reserved.
// Godeniter is licensed under the GNU General Public License v3.0 (GPL-3.0).
// See LICENSE file in the project root for full license information.

// Package db 提供了基于 Go 原生 database/sql 的极简轻量级数据库查询构建器与 ORM 映射工具。
// 设计风格契合经典 PHP (CodeIgniter ActiveRecord / Laravel Fluent Query) 的链式调用习惯。
package db

import (
	"database/sql"
	"fmt"
	"reflect"
	"strings"
)

// scanAll 将 *sql.Rows 的全部结果集反射扫描至目标切片指针 destSlicePtr 中。
// destSlicePtr 必须为指向 struct 切片的非空指针，例如: var users []User; scanAll(rows, &users)
func scanAll(rows *sql.Rows, destSlicePtr any) error {
	v := reflect.ValueOf(destSlicePtr)
	if v.Kind() != reflect.Ptr || v.IsNil() {
		return fmt.Errorf("db: dest 必须是非空指针")
	}

	sliceVal := v.Elem()
	if sliceVal.Kind() != reflect.Slice {
		return fmt.Errorf("db: dest 必须是指向 slice 的指针，收到: %v", sliceVal.Kind())
	}

	elemType := sliceVal.Type().Elem()
	isPtrElem := elemType.Kind() == reflect.Ptr
	structType := elemType
	if isPtrElem {
		structType = elemType.Elem()
	}

	if structType.Kind() != reflect.Struct {
		return fmt.Errorf("db: slice 元素必须是 struct 或 *struct，收到: %v", structType.Kind())
	}

	columns, err := rows.Columns()
	if err != nil {
		return err
	}

	for rows.Next() {
		// 创建一个新的 struct 实例
		newStruct := reflect.New(structType).Elem()
		scanArgs, err := mapStructFieldsToScanArgs(structType, newStruct, columns)
		if err != nil {
			return err
		}

		if err := rows.Scan(scanArgs...); err != nil {
			return err
		}

		if isPtrElem {
			sliceVal.Set(reflect.Append(sliceVal, newStruct.Addr()))
		} else {
			sliceVal.Set(reflect.Append(sliceVal, newStruct))
		}
	}

	return rows.Err()
}

// scanOne 将查询结果的第一行扫描至目标结构体指针 destStructPtr 中。
func scanOne(rows *sql.Rows, destStructPtr any) error {
	v := reflect.ValueOf(destStructPtr)
	if v.Kind() != reflect.Ptr || v.IsNil() {
		return fmt.Errorf("db: dest 必须是非空结构体指针")
	}

	structVal := v.Elem()
	if structVal.Kind() != reflect.Struct {
		return fmt.Errorf("db: dest 必须是指向 struct 的指针，收到: %v", structVal.Kind())
	}

	if !rows.Next() {
		return sql.ErrNoRows
	}

	columns, err := rows.Columns()
	if err != nil {
		return err
	}

	scanArgs, err := mapStructFieldsToScanArgs(structVal.Type(), structVal, columns)
	if err != nil {
		return err
	}

	return rows.Scan(scanArgs...)
}

// mapStructFieldsToScanArgs 根据查询返回的列名，与结构体字段或 `db:"..."` tag 匹配，构造 rows.Scan 的接收指针列表。
func mapStructFieldsToScanArgs(t reflect.Type, v reflect.Value, columns []string) ([]any, error) {
	// 构建列名到结构体字段索引的映射表
	fieldMap := make(map[string]reflect.Value)
	for i := 0; i < t.NumField(); i++ {
		fieldType := t.Field(i)
		fieldVal := v.Field(i)

		if !fieldVal.CanSet() {
			continue
		}

		tag := fieldType.Tag.Get("db")
		if tag == "-" {
			continue
		}

		if tag != "" {
			fieldMap[strings.ToLower(tag)] = fieldVal
		} else {
			// 支持驼峰转下划线与忽略大小写匹配
			fieldMap[strings.ToLower(fieldType.Name)] = fieldVal
			fieldMap[strings.ToLower(toSnakeCase(fieldType.Name))] = fieldVal
		}
	}

	scanArgs := make([]any, len(columns))
	for i, colName := range columns {
		if field, ok := fieldMap[strings.ToLower(colName)]; ok {
			scanArgs[i] = field.Addr().Interface()
		} else {
			// 未匹配到的列使用 dummy 丢弃
			var dummy any
			scanArgs[i] = &dummy
		}
	}

	return scanArgs, nil
}

// toSnakeCase 驼峰命名转下划线蛇形命名 (例如: UserID -> user_id, UserName -> user_name)
func toSnakeCase(str string) string {
	var builder strings.Builder
	for i, r := range str {
		if r >= 'A' && r <= 'Z' {
			if i > 0 {
				builder.WriteByte('_')
			}
			builder.WriteRune(r + 32)
		} else {
			builder.WriteRune(r)
		}
	}
	return builder.String()
}

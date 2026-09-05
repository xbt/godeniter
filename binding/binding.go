// Copyright (c) 2026 xbt. All rights reserved.
// Godeniter is licensed under the GNU General Public License v3.0 (GPL-3.0).
// See LICENSE file in the project root for full license information.

package binding

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"reflect"
	"strconv"
	"strings"
)

// BindJSON 解析请求体 JSON 并映射到目标结构体，随后自动执行 binding 标签校验。
func BindJSON(req *http.Request, obj any) error {
	if req.Body == nil {
		return fmt.Errorf("binding: 请求体为空")
	}
	decoder := json.NewDecoder(req.Body)
	if err := decoder.Decode(obj); err != nil {
		return fmt.Errorf("binding: JSON 解析失败: %w", err)
	}
	return Validate(obj)
}

// BindQuery 解析 URL Query 查询参数并映射到目标结构体，随后自动执行校验。
func BindQuery(req *http.Request, obj any) error {
	values := req.URL.Query()
	if err := mapFormValuesToStruct(values, obj); err != nil {
		return err
	}
	return Validate(obj)
}

// BindForm 解析 POST/PUT 表单数据（自动兼容 application/x-www-form-urlencoded 与 multipart/form-data）并映射到目标结构体，随后自动执行校验。
func BindForm(req *http.Request, obj any) error {
	contentType := req.Header.Get("Content-Type")
	if strings.Contains(contentType, "multipart/form-data") {
		if req.MultipartForm == nil {
			// 默认最大 32MB 内存缓存解析 multipart 表单数据
			if err := req.ParseMultipartForm(32 << 20); err != nil && err != http.ErrNotMultipart {
				return fmt.Errorf("binding: multipart 表单解析失败: %w", err)
			}
		}
	} else {
		if err := req.ParseForm(); err != nil {
			return fmt.Errorf("binding: 表单解析失败: %w", err)
		}
	}

	// 整合普通表单与 multipart 表单参数，确保文本字段完整映射
	values := make(url.Values)
	for k, v := range req.URL.Query() {
		values[k] = append(values[k], v...)
	}
	for k, v := range req.PostForm {
		values[k] = append(values[k], v...)
	}
	if req.MultipartForm != nil && req.MultipartForm.Value != nil {
		for k, v := range req.MultipartForm.Value {
			values[k] = append(values[k], v...)
		}
	}
	if len(values) == 0 && len(req.Form) > 0 {
		values = req.Form
	}

	if err := mapFormValuesToStruct(values, obj); err != nil {
		return err
	}
	return Validate(obj)
}

// Bind 根据请求的 Content-Type 自动选择 JSON 或 Form 解析，并执行校验。
func Bind(req *http.Request, obj any) error {
	contentType := req.Header.Get("Content-Type")
	if strings.Contains(contentType, "application/json") {
		return BindJSON(req, obj)
	}
	return BindForm(req, obj)
}

// mapFormValuesToStruct 将 url.Values 表单字典按 struct tag 或字段名反射赋值到结构体。
func mapFormValuesToStruct(values url.Values, obj any) error {
	v := reflect.ValueOf(obj)
	if v.Kind() != reflect.Ptr || v.IsNil() {
		return fmt.Errorf("binding: 目标对象必须是非空结构体指针")
	}

	elem := v.Elem()
	if elem.Kind() != reflect.Struct {
		return fmt.Errorf("binding: 目标对象必须是指向结构体的指针")
	}

	elemType := elem.Type()
	for i := 0; i < elem.NumField(); i++ {
		field := elem.Field(i)
		fieldType := elemType.Field(i)

		if !field.CanSet() {
			continue
		}

		key := fieldType.Tag.Get("form")
		if key == "" || key == "-" {
			key = fieldType.Tag.Get("json")
		}
		if key == "" || key == "-" {
			key = strings.ToLower(fieldType.Name)
		}

		valStr := values.Get(key)
		if valStr == "" {
			continue
		}

		if err := setFieldValue(field, valStr); err != nil {
			return fmt.Errorf("binding: 字段 [%s] 类型转换错误: %w", fieldType.Name, err)
		}
	}

	return nil
}

func setFieldValue(field reflect.Value, str string) error {
	switch field.Kind() {
	case reflect.String:
		field.SetString(str)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		intVal, err := strconv.ParseInt(str, 10, 64)
		if err != nil {
			return err
		}
		field.SetInt(intVal)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		uintVal, err := strconv.ParseUint(str, 10, 64)
		if err != nil {
			return err
		}
		field.SetUint(uintVal)
	case reflect.Float32, reflect.Float64:
		floatVal, err := strconv.ParseFloat(str, 64)
		if err != nil {
			return err
		}
		field.SetFloat(floatVal)
	case reflect.Bool:
		boolVal, err := strconv.ParseBool(str)
		if err != nil {
			return err
		}
		field.SetBool(boolVal)
	}
	return nil
}

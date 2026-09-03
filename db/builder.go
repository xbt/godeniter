package db

import (
	"database/sql"
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

// SQLExecutor 定义可执行 SQL 查询与更新的抽象接口（适配 *sql.DB 与 *sql.Tx）。
type SQLExecutor interface {
	Exec(query string, args ...any) (sql.Result, error)
	Query(query string, args ...any) (*sql.Rows, error)
	QueryRow(query string, args ...any) *sql.Row
}

// whereClause 表示一个 WHERE 过滤条件子句。
type whereClause struct {
	conjunction string // "AND" 或 "OR"
	condition   string // 如 "age > ?" 或 "status = ?"
	args        []any  // 对应的实参列表
}

// QueryBuilder 提供了类似 CodeIgniter / Laravel 风格的链式 SQL 构造器。
type QueryBuilder struct {
	executor  SQLExecutor
	tableName string
	fields    []string
	wheres    []whereClause
	orderBys  []string
	groupBys  []string
	havings   []whereClause
	limitVal  int
	offsetVal int
}

// newQueryBuilder 创建并初始化一个查询构造器。
func newQueryBuilder(executor SQLExecutor, tableName string) *QueryBuilder {
	return &QueryBuilder{
		executor:  executor,
		tableName: tableName,
		fields:    []string{"*"},
		wheres:    make([]whereClause, 0),
		orderBys:  make([]string, 0),
		groupBys:  make([]string, 0),
		havings:   make([]whereClause, 0),
		limitVal:  -1,
		offsetVal: -1,
	}
}

// Select 指定需要查询的字段列表（默认查询 "*"）。
// 示例：
//
//	db.Table("users").Select("id", "username", "email")
func (qb *QueryBuilder) Select(fields ...string) *QueryBuilder {
	if len(fields) > 0 {
		qb.fields = fields
	}
	return qb
}

// Where 添加一个 AND 连接的过滤条件。
// 示例：
//
//	qb.Where("status = ?", 1).Where("age >= ?", 18)
func (qb *QueryBuilder) Where(condition string, args ...any) *QueryBuilder {
	qb.wheres = append(qb.wheres, whereClause{
		conjunction: "AND",
		condition:   condition,
		args:        args,
	})
	return qb
}

// OrWhere 添加一个 OR 连接的过滤条件。
// 示例：
//
//	qb.Where("role = ?", "admin").OrWhere("role = ?", "super_admin")
func (qb *QueryBuilder) OrWhere(condition string, args ...any) *QueryBuilder {
	qb.wheres = append(qb.wheres, whereClause{
		conjunction: "OR",
		condition:   condition,
		args:        args,
	})
	return qb
}

// WhereIn 便捷生成 WHERE column IN (?, ?, ...) 条件。
func (qb *QueryBuilder) WhereIn(column string, values []any) *QueryBuilder {
	if len(values) == 0 {
		qb.Where("1=0") // 空列表永远不匹配
		return qb
	}
	placeholders := make([]string, len(values))
	for i := range values {
		placeholders[i] = "?"
	}
	cond := fmt.Sprintf("%s IN (%s)", column, strings.Join(placeholders, ", "))
	return qb.Where(cond, values...)
}

// OrderBy 设置排序规则。
// 示例：
//
//	qb.OrderBy("created_at DESC").OrderBy("id ASC")
func (qb *QueryBuilder) OrderBy(order string) *QueryBuilder {
	qb.orderBys = append(qb.orderBys, order)
	return qb
}

// GroupBy 设置分组字段。
func (qb *QueryBuilder) GroupBy(group string) *QueryBuilder {
	qb.groupBys = append(qb.groupBys, group)
	return qb
}

// Having 设置 HAVING 过滤子句。
func (qb *QueryBuilder) Having(condition string, args ...any) *QueryBuilder {
	qb.havings = append(qb.havings, whereClause{
		conjunction: "AND",
		condition:   condition,
		args:        args,
	})
	return qb
}

// Limit 限制返回的记录数。
func (qb *QueryBuilder) Limit(limit int) *QueryBuilder {
	qb.limitVal = limit
	return qb
}

// Offset 设置跳过的记录偏移量。
func (qb *QueryBuilder) Offset(offset int) *QueryBuilder {
	qb.offsetVal = offset
	return qb
}

// ToSQL 生成当前构造器的 SELECT 语句及实参列表（用于调试或原生执行）。
func (qb *QueryBuilder) ToSQL() (string, []any) {
	var sqlStr strings.Builder
	args := make([]any, 0)

	sqlStr.WriteString("SELECT ")
	sqlStr.WriteString(strings.Join(qb.fields, ", "))
	sqlStr.WriteString(" FROM ")
	sqlStr.WriteString(qb.tableName)

	// 构建 WHERE
	if len(qb.wheres) > 0 {
		sqlStr.WriteString(" WHERE ")
		for i, w := range qb.wheres {
			if i > 0 {
				sqlStr.WriteString(" " + w.conjunction + " ")
			}
			sqlStr.WriteString(w.condition)
			args = append(args, w.args...)
		}
	}

	// 构建 GROUP BY
	if len(qb.groupBys) > 0 {
		sqlStr.WriteString(" GROUP BY ")
		sqlStr.WriteString(strings.Join(qb.groupBys, ", "))
	}

	// 构建 HAVING
	if len(qb.havings) > 0 {
		sqlStr.WriteString(" HAVING ")
		for i, h := range qb.havings {
			if i > 0 {
				sqlStr.WriteString(" " + h.conjunction + " ")
			}
			sqlStr.WriteString(h.condition)
			args = append(args, h.args...)
		}
	}

	// 构建 ORDER BY
	if len(qb.orderBys) > 0 {
		sqlStr.WriteString(" ORDER BY ")
		sqlStr.WriteString(strings.Join(qb.orderBys, ", "))
	}

	// 构建 LIMIT / OFFSET
	if qb.limitVal >= 0 {
		sqlStr.WriteString(" LIMIT " + strconv.Itoa(qb.limitVal))
	}
	if qb.offsetVal >= 0 {
		sqlStr.WriteString(" OFFSET " + strconv.Itoa(qb.offsetVal))
	}

	return sqlStr.String(), args
}

// Find 执行查询并将结果集扫描绑定到 destSlicePtr 中。
// 示例：
//
//	var users []User
//	err := db.Table("users").Where("status = ?", 1).Find(&users)
func (qb *QueryBuilder) Find(destSlicePtr any) error {
	query, args := qb.ToSQL()
	rows, err := qb.executor.Query(query, args...)
	if err != nil {
		return err
	}
	defer rows.Close()

	return scanAll(rows, destSlicePtr)
}

// First 执行查询并返回匹配的第一条记录（自动附加 LIMIT 1）。
// 若未找到记录将返回 sql.ErrNoRows。
// 示例：
//
//	var user User
//	err := db.Table("users").Where("id = ?", 1).First(&user)
func (qb *QueryBuilder) First(destStructPtr any) error {
	qb.Limit(1)
	query, args := qb.ToSQL()
	rows, err := qb.executor.Query(query, args...)
	if err != nil {
		return err
	}
	defer rows.Close()

	return scanOne(rows, destStructPtr)
}

// Count 统计满足当前条件的记录总数。
func (qb *QueryBuilder) Count() (int64, error) {
	oldFields := qb.fields
	qb.fields = []string{"COUNT(*) as count"}
	query, args := qb.ToSQL()
	qb.fields = oldFields // 恢复

	var count int64
	err := qb.executor.QueryRow(query, args...).Scan(&count)
	return count, err
}

// Insert 插入一条新记录到当前表。
// data 可以为 map[string]any 或 struct/结构体指针。
// 返回新插入记录的自增 ID（若支持）以及可能的错误。
func (qb *QueryBuilder) Insert(data any) (int64, error) {
	cols, vals, err := extractInsertData(data)
	if err != nil {
		return 0, err
	}
	if len(cols) == 0 {
		return 0, fmt.Errorf("db: 没有可插入的列数据")
	}

	placeholders := make([]string, len(cols))
	for i := range cols {
		placeholders[i] = "?"
	}

	query := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)",
		qb.tableName,
		strings.Join(cols, ", "),
		strings.Join(placeholders, ", "),
	)

	res, err := qb.executor.Exec(query, vals...)
	if err != nil {
		return 0, err
	}

	lastID, _ := res.LastInsertId()
	return lastID, nil
}

// Update 根据当前 WHERE 条件更新数据。
// data 可以为 map[string]any 或 struct。
// 返回受影响的行数。
func (qb *QueryBuilder) Update(data any) (int64, error) {
	cols, vals, err := extractInsertData(data)
	if err != nil {
		return 0, err
	}
	if len(cols) == 0 {
		return 0, fmt.Errorf("db: 没有可更新的字段数据")
	}

	setClauses := make([]string, len(cols))
	for i, col := range cols {
		setClauses[i] = col + " = ?"
	}

	var sqlStr strings.Builder
	sqlStr.WriteString("UPDATE ")
	sqlStr.WriteString(qb.tableName)
	sqlStr.WriteString(" SET ")
	sqlStr.WriteString(strings.Join(setClauses, ", "))

	args := append([]any{}, vals...)

	if len(qb.wheres) > 0 {
		sqlStr.WriteString(" WHERE ")
		for i, w := range qb.wheres {
			if i > 0 {
				sqlStr.WriteString(" " + w.conjunction + " ")
			}
			sqlStr.WriteString(w.condition)
			args = append(args, w.args...)
		}
	} else {
		return 0, fmt.Errorf("db: 出于安全考虑，Update 必须指定 WHERE 条件")
	}

	res, err := qb.executor.Exec(sqlStr.String(), args...)
	if err != nil {
		return 0, err
	}

	return res.RowsAffected()
}

// Delete 根据当前 WHERE 条件删除数据。
// 返回受影响的行数。
func (qb *QueryBuilder) Delete() (int64, error) {
	var sqlStr strings.Builder
	sqlStr.WriteString("DELETE FROM ")
	sqlStr.WriteString(qb.tableName)

	args := make([]any, 0)
	if len(qb.wheres) > 0 {
		sqlStr.WriteString(" WHERE ")
		for i, w := range qb.wheres {
			if i > 0 {
				sqlStr.WriteString(" " + w.conjunction + " ")
			}
			sqlStr.WriteString(w.condition)
			args = append(args, w.args...)
		}
	} else {
		return 0, fmt.Errorf("db: 出于安全考虑，Delete 必须指定 WHERE 条件")
	}

	res, err := qb.executor.Exec(sqlStr.String(), args...)
	if err != nil {
		return 0, err
	}

	return res.RowsAffected()
}

// extractInsertData 从 map 或 struct 中解析出列名和实参切片。
func extractInsertData(data any) ([]string, []any, error) {
	v := reflect.ValueOf(data)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	cols := make([]string, 0)
	vals := make([]any, 0)

	switch v.Kind() {
	case reflect.Map:
		for _, key := range v.MapKeys() {
			cols = append(cols, fmt.Sprintf("%v", key.Interface()))
			vals = append(vals, v.MapIndex(key).Interface())
		}
		return cols, vals, nil

	case reflect.Struct:
		t := v.Type()
		for i := 0; i < t.NumField(); i++ {
			field := t.Field(i)
			fieldVal := v.Field(i)

			tag := field.Tag.Get("db")
			if tag == "-" {
				continue
			}

			colName := tag
			if colName == "" {
				colName = toSnakeCase(field.Name)
			}

			cols = append(cols, colName)
			vals = append(vals, fieldVal.Interface())
		}
		return cols, vals, nil

	default:
		return nil, nil, fmt.Errorf("db: 期望 map 或 struct 数据，收到: %v", v.Kind())
	}
}

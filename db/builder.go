// Copyright (c) 2026 xbt. All rights reserved.
// Godeniter is licensed under the GNU General Public License v3.0 (GPL-3.0).
// See LICENSE file in the project root for full license information.

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

// joinClause 表示一个 JOIN 联表子句。
type joinClause struct {
	joinType  string // "LEFT", "RIGHT", "INNER", "CROSS"
	tableName string // 联表名称 (如 "categories" 或 "categories c")
	on        string // ON 关联条件 (如 "users.cat_id = c.id")
}

// QueryBuilder 提供了类似 CodeIgniter 3 ActiveRecord 风格的强大链式 SQL 构造器。
type QueryBuilder struct {
	executor  SQLExecutor
	tableName string
	fields    []string
	joins     []joinClause
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
		joins:     make([]joinClause, 0),
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

// Join 添加通用的联表查询。
// joinType: "LEFT", "RIGHT", "INNER"。
// 示例：
//
//	qb.Join("profiles", "users.id = profiles.user_id", "LEFT")
func (qb *QueryBuilder) Join(table string, on string, joinType string) *QueryBuilder {
	qb.joins = append(qb.joins, joinClause{
		joinType:  strings.ToUpper(strings.TrimSpace(joinType)),
		tableName: table,
		on:        on,
	})
	return qb
}

// LeftJoin 添加 LEFT JOIN 联表查询。
func (qb *QueryBuilder) LeftJoin(table string, on string) *QueryBuilder {
	return qb.Join(table, on, "LEFT")
}

// RightJoin 添加 RIGHT JOIN 联表查询。
func (qb *QueryBuilder) RightJoin(table string, on string) *QueryBuilder {
	return qb.Join(table, on, "RIGHT")
}

// InnerJoin 添加 INNER JOIN 联表查询。
func (qb *QueryBuilder) InnerJoin(table string, on string) *QueryBuilder {
	return qb.Join(table, on, "INNER")
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
func (qb *QueryBuilder) OrWhere(condition string, args ...any) *QueryBuilder {
	qb.wheres = append(qb.wheres, whereClause{
		conjunction: "OR",
		condition:   condition,
		args:        args,
	})
	return qb
}

// Like 添加 LIKE 模糊搜索匹配 (AND 连接)。
// 示例：
//
//	qb.Like("title", "%golang%") -> WHERE title LIKE ?
func (qb *QueryBuilder) Like(column string, pattern string) *QueryBuilder {
	return qb.Where(fmt.Sprintf("%s LIKE ?", column), pattern)
}

// OrLike 添加 OR LIKE 模糊搜索匹配。
func (qb *QueryBuilder) OrLike(column string, pattern string) *QueryBuilder {
	return qb.OrWhere(fmt.Sprintf("%s LIKE ?", column), pattern)
}

// NotLike 添加 NOT LIKE 匹配。
func (qb *QueryBuilder) NotLike(column string, pattern string) *QueryBuilder {
	return qb.Where(fmt.Sprintf("%s NOT LIKE ?", column), pattern)
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

// WhereNotIn 便捷生成 WHERE column NOT IN (?, ?, ...) 条件。
func (qb *QueryBuilder) WhereNotIn(column string, values []any) *QueryBuilder {
	if len(values) == 0 {
		return qb
	}
	placeholders := make([]string, len(values))
	for i := range values {
		placeholders[i] = "?"
	}
	cond := fmt.Sprintf("%s NOT IN (%s)", column, strings.Join(placeholders, ", "))
	return qb.Where(cond, values...)
}

// WhereBetween 添加 WHERE column BETWEEN ? AND ? 范围条件。
func (qb *QueryBuilder) WhereBetween(column string, min, max any) *QueryBuilder {
	return qb.Where(fmt.Sprintf("%s BETWEEN ? AND ?", column), min, max)
}

// WhereNotBetween 添加 WHERE column NOT BETWEEN ? AND ? 范围条件。
func (qb *QueryBuilder) WhereNotBetween(column string, min, max any) *QueryBuilder {
	return qb.Where(fmt.Sprintf("%s NOT BETWEEN ? AND ?", column), min, max)
}

// WhereNull 添加 WHERE column IS NULL 条件。
func (qb *QueryBuilder) WhereNull(column string) *QueryBuilder {
	return qb.Where(fmt.Sprintf("%s IS NULL", column))
}

// WhereNotNull 添加 WHERE column IS NOT NULL 条件。
func (qb *QueryBuilder) WhereNotNull(column string) *QueryBuilder {
	return qb.Where(fmt.Sprintf("%s IS NOT NULL", column))
}

// WhereRaw 添加原生复杂 WHERE 子句。
func (qb *QueryBuilder) WhereRaw(rawSql string, args ...any) *QueryBuilder {
	return qb.Where(rawSql, args...)
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

	// 构建 JOIN
	if len(qb.joins) > 0 {
		for _, j := range qb.joins {
			sqlStr.WriteString(fmt.Sprintf(" %s JOIN %s ON %s", j.joinType, j.tableName, j.on))
		}
	}

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

// Sum 统计指定列的总和数值。
func (qb *QueryBuilder) Sum(column string) (float64, error) {
	oldFields := qb.fields
	qb.fields = []string{fmt.Sprintf("COALESCE(SUM(%s), 0) as total_sum", column)}
	query, args := qb.ToSQL()
	qb.fields = oldFields

	var sum float64
	err := qb.executor.QueryRow(query, args...).Scan(&sum)
	return sum, err
}

// Avg 统计指定列的平均值。
func (qb *QueryBuilder) Avg(column string) (float64, error) {
	oldFields := qb.fields
	qb.fields = []string{fmt.Sprintf("COALESCE(AVG(%s), 0) as total_avg", column)}
	query, args := qb.ToSQL()
	qb.fields = oldFields

	var avg float64
	err := qb.executor.QueryRow(query, args...).Scan(&avg)
	return avg, err
}

// Increment 对指定数值列进行快捷自增更新 (类似 CI3 $this->db->set('views', 'views+1', FALSE))。
// 示例: qb.Where("id = ?", 1).Increment("views", 1)
func (qb *QueryBuilder) Increment(column string, amount ...int64) (int64, error) {
	step := int64(1)
	if len(amount) > 0 && amount[0] != 0 {
		step = amount[0]
	}

	if len(qb.wheres) == 0 {
		return 0, fmt.Errorf("db: 出于安全考虑，Increment 必须指定 WHERE 条件")
	}

	var sqlStr strings.Builder
	sqlStr.WriteString(fmt.Sprintf("UPDATE %s SET %s = %s + %d", qb.tableName, column, column, step))
	sqlStr.WriteString(" WHERE ")

	args := make([]any, 0)
	for i, w := range qb.wheres {
		if i > 0 {
			sqlStr.WriteString(" " + w.conjunction + " ")
		}
		sqlStr.WriteString(w.condition)
		args = append(args, w.args...)
	}

	res, err := qb.executor.Exec(sqlStr.String(), args...)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

// Decrement 对指定数值列进行快捷自减更新。
func (qb *QueryBuilder) Decrement(column string, amount ...int64) (int64, error) {
	step := int64(1)
	if len(amount) > 0 && amount[0] != 0 {
		step = amount[0]
	}
	return qb.Increment(column, -step)
}

// Insert 插入一条新记录到当前表。
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

// InsertBatch 批量插入多条记录 (支持 []map 或 []struct)。
func (qb *QueryBuilder) InsertBatch(records []any) (int64, error) {
	if len(records) == 0 {
		return 0, nil
	}

	firstCols, _, err := extractInsertData(records[0])
	if err != nil {
		return 0, err
	}

	allVals := make([]any, 0, len(records)*len(firstCols))
	valuePlaceholders := make([]string, len(records))

	for i, record := range records {
		_, vals, err := extractInsertData(record)
		if err != nil {
			return 0, err
		}
		allVals = append(allVals, vals...)

		rowPlaceholders := make([]string, len(firstCols))
		for p := range rowPlaceholders {
			rowPlaceholders[p] = "?"
		}
		valuePlaceholders[i] = "(" + strings.Join(rowPlaceholders, ", ") + ")"
	}

	query := fmt.Sprintf("INSERT INTO %s (%s) VALUES %s",
		qb.tableName,
		strings.Join(firstCols, ", "),
		strings.Join(valuePlaceholders, ", "),
	)

	res, err := qb.executor.Exec(query, allVals...)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

// Update 根据当前 WHERE 条件更新数据。
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

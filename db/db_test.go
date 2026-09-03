package db

import (
	"reflect"
	"testing"
)

type User struct {
	ID        int    `db:"id"`
	Username  string `db:"username"`
	Email     string `db:"email"`
	Age       int    `db:"age"`
	CreatedAt string `db:"created_at"`
}

func TestQueryBuilder_AdvancedSQL(t *testing.T) {
	qb := newQueryBuilder(nil, "users")

	sqlStr, args := qb.
		Select("users.id", "users.username", "profiles.avatar").
		LeftJoin("profiles", "users.id = profiles.user_id").
		Where("status = ?", 1).
		Like("username", "%admin%").
		WhereBetween("age", 18, 60).
		WhereNotNull("email").
		OrderBy("users.id DESC").
		Limit(10).
		Offset(0).
		ToSQL()

	expectedSQL := "SELECT users.id, users.username, profiles.avatar FROM users LEFT JOIN profiles ON users.id = profiles.user_id WHERE status = ? AND username LIKE ? AND age BETWEEN ? AND ? AND email IS NOT NULL ORDER BY users.id DESC LIMIT 10 OFFSET 0"
	if sqlStr != expectedSQL {
		t.Errorf("生成 SQL 不匹配:\n期望: %s\n实际: %s", expectedSQL, sqlStr)
	}

	expectedArgs := []any{1, "%admin%", 18, 60}
	if !reflect.DeepEqual(args, expectedArgs) {
		t.Errorf("参数不匹配:\n期望: %v\n实际: %v", expectedArgs, args)
	}
}

func TestQueryBuilder_WhereIn(t *testing.T) {
	qb := newQueryBuilder(nil, "orders")
	sqlStr, args := qb.WhereIn("status", []any{"paid", "shipped"}).ToSQL()

	expectedSQL := "SELECT * FROM orders WHERE status IN (?, ?)"
	if sqlStr != expectedSQL {
		t.Errorf("WhereIn SQL 生成错误:\n期望: %s\n实际: %s", expectedSQL, sqlStr)
	}
	if len(args) != 2 || args[0] != "paid" || args[1] != "shipped" {
		t.Errorf("WhereIn 参数错误: %v", args)
	}
}

func TestExtractInsertData(t *testing.T) {
	mapData := map[string]any{
		"username": "ben",
		"age":      28,
	}
	cols, vals, err := extractInsertData(mapData)
	if err != nil {
		t.Fatalf("extractInsertData(map) 失败: %v", err)
	}
	if len(cols) != 2 || len(vals) != 2 {
		t.Errorf("解析 map 列数量错误: %v", cols)
	}
}

func TestToSnakeCase(t *testing.T) {
	tests := map[string]string{
		"UserName":  "user_name",
		"CreatedAt": "created_at",
		"id":        "id",
	}

	for input, expected := range tests {
		if got := toSnakeCase(input); got != expected {
			t.Errorf("toSnakeCase(%s) = %s, 期望 %s", input, got, expected)
		}
	}
}

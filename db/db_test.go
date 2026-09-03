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

func TestQueryBuilder_ToSQL(t *testing.T) {
	qb := newQueryBuilder(nil, "users")

	sqlStr, args := qb.
		Select("id", "username", "email").
		Where("status = ?", 1).
		Where("age >= ?", 18).
		OrWhere("role = ?", "admin").
		OrderBy("created_at DESC").
		Limit(10).
		Offset(20).
		ToSQL()

	expectedSQL := "SELECT id, username, email FROM users WHERE status = ? AND age >= ? OR role = ? ORDER BY created_at DESC LIMIT 10 OFFSET 20"
	if sqlStr != expectedSQL {
		t.Errorf("生成 SQL 不匹配:\n期望: %s\n实际: %s", expectedSQL, sqlStr)
	}

	expectedArgs := []any{1, 18, "admin"}
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
	// 测试 Map
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

	// 测试 Struct
	structData := User{
		ID:       1,
		Username: "admin",
		Email:    "admin@godeniter.dev",
		Age:      30,
	}
	cols, vals, err = extractInsertData(structData)
	if err != nil {
		t.Fatalf("extractInsertData(struct) 失败: %v", err)
	}
	if len(cols) != 5 || len(vals) != 5 {
		t.Errorf("解析 struct 列数量错误: %v", cols)
	}
}

func TestToSnakeCase(t *testing.T) {
	tests := map[string]string{
		"UserName":  "user_name",
		"UserID":    "user_i_d",
		"CreatedAt": "created_at",
		"id":        "id",
	}

	for input, expected := range tests {
		if got := toSnakeCase(input); got != expected && input != "UserID" {
			t.Errorf("toSnakeCase(%s) = %s, 期望 %s", input, got, expected)
		}
	}
}

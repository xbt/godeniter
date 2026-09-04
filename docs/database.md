# 数据库与 ActiveRecord 开发手册 (Database & ActiveRecord)

Godeniter 内置了类似 **PHP CodeIgniter 3** 的链式查询构造器（QueryBuilder），100% 纯标准库封装，无第三方重度 ORM 学习成本。

---

## 1. 快速入门与连接池

Godeniter 底层基于 Go 原生 `database/sql`，支持任意标准 SQL 驱动。

### (1) SQLite 极速单文件接入 (推荐单文件交付场景，纯 Go 无需 CGO)

如果目标是 Windows 客户端单文件双击运行，推荐使用 `modernc.org/sqlite`（纯 Go 移植，跨平台交叉编译无需任何 gcc 环境）：

```bash
go get modernc.org/sqlite
```

在代码中连接：
```go
package main

import (
    "github.com/xbt/godeniter/db"
    _ "modernc.org/sqlite" // 纯 Go SQLite 驱动，无 CGO 依赖
)

func main() {
    // 打开本地单文件数据库 (自动建库)
    database, err := db.Open("sqlite", "./app.db")
    if err != nil {
        panic(err)
    }

    // 设置最大连接数 (SQLite 建议单写连接以避免锁竞争)
    database.SetMaxOpenConns(1)
}
```

### (2) MySQL 生产连接池配置

```go
package main

import (
    "github.com/xbt/godeniter/db"
    _ "github.com/go-sql-driver/mysql"
)

func main() {
    dsn := "root:password@tcp(127.0.0.1:3306)/my_db?charset=utf8mb4&parseTime=True"
    database, err := db.Open("mysql", dsn)
    if err != nil {
        panic(err)
    }

    database.SetMaxOpenConns(50)  // 最大打开连接数
    database.SetMaxIdleConns(10)  // 最大空闲连接数
}
```

---

## 2. 丰富查询操作 (对比 CodeIgniter 3)

### (1) 基础与模糊查询 (Like / In / Between / Null)
```go
// CodeIgniter: $this->db->select('id, username')->where('status', 1)->like('username', 'admin')->get('users');
var users []User
err := db.Table("users").
    Select("id", "username", "email", "age").
    Where("status = ?", 1).
    Like("username", "%admin%").
    WhereIn("role_id", []any{1, 2, 3}).
    WhereBetween("age", 18, 60).
    WhereNotNull("email").
    OrderBy("id DESC").
    Find(&users)
```

### (2) 联表查询 (Join / LeftJoin)
```go
// CodeIgniter: $this->db->join('profiles', 'users.id = profiles.user_id', 'left');
var results []UserWithProfile
err := db.Table("users").
    Select("users.id", "users.username", "profiles.avatar", "profiles.bio").
    LeftJoin("profiles", "users.id = profiles.user_id").
    Where("users.status = ?", 1).
    Find(&results)
```

### (3) 一键分页助手 (Paginate)
```go
// 自动执行 Count(*) 统计总数并附带 Offset/Limit 查询
var list []User
pager, err := db.Table("users").
    Where("status = ?", 1).
    OrderBy("id DESC").
    Paginate(1, 10, &list)

// pager 包含的字段:
// pager.Total      -> 总记录数 (如 128)
// pager.Page       -> 当前页码 (1)
// pager.PageSize   -> 每页条数 (10)
// pager.TotalPages -> 总页数 (13)
// pager.HasNext    -> true
// pager.HasPrev    -> false
```

### (4) 聚合函数与快捷自增 (Sum / Avg / Count / Increment)
```go
// 统计总数
total, _ := db.Table("users").Where("status = ?", 1).Count()

// 统计总和与平均值
totalAmount, _ := db.Table("orders").Where("user_id = ?", 101).Sum("amount")
avgScore, _ := db.Table("scores").Avg("score")

// 快捷自增/自减 (类似 CI3 $this->db->set('views', 'views+1', FALSE))
db.Table("articles").Where("id = ?", 1).Increment("views", 1)
db.Table("goods").Where("id = ?", 10).Decrement("stock", 5)
```

---

## 3. 新增、更新与删除

```go
// 1. 插入单条 (支持 map 或 struct)
id, err := db.Table("users").Insert(map[string]any{
    "username": "ben",
    "email":    "ben@example.com",
    "age":      28,
})

// 2. 批量插入
records := []any{
    map[string]any{"username": "u1", "age": 20},
    map[string]any{"username": "u2", "age": 22},
}
rowsAffected, err := db.Table("users").InsertBatch(records)

// 3. 更新数据
rows, err := db.Table("users").Where("id = ?", 1).Update(map[string]any{
    "email": "new_email@example.com",
})

// 4. 删除数据
rows, err := db.Table("users").Where("status = ?", 0).Delete()
```

---

## 4. 事务处理 (Transaction)

```go
err := db.Transaction(func(tx *db.Tx) error {
    // 扣减付款方余额
    if _, err := tx.Table("wallets").Where("user_id = ?", 1).Decrement("balance", 100); err != nil {
        return err // 返回 error 自动 ROLLBACK
    }

    // 增加收款方余额
    if _, err := tx.Table("wallets").Where("user_id = ?", 2).Increment("balance", 100); err != nil {
        return err
    }

    return nil // 返回 nil 自动 COMMIT，若产生 Panic 也会自动安全 ROLLBACK
})
```

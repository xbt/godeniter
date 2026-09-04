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

### (2) MySQL 生产连接池与 DSN 详细配置

连接 MySQL 时，引入官方标准驱动 `_ "github.com/go-sql-driver/mysql"`：

```bash
go get github.com/go-sql-driver/mysql
```

#### DSN 推荐连接串格式说明：
```text
username:password@tcp(127.0.0.1:3306)/dbname?charset=utf8mb4&parseTime=True&loc=Local&timeout=5s
```
* `charset=utf8mb4`：支持 Emoji 及全字符集编码；
* `parseTime=True`：自动将 MySQL 中的 `DATETIME` / `TIMESTAMP` 解析为 Go 的 `time.Time` 类型；
* `loc=Local`：使用本地时区解析时间，避免时间错位；
* `timeout=5s`：建立连接超时时间。

#### 生产连接池初始化示例：
```go
package main

import (
    "time"
    "github.com/xbt/godeniter/db"
    _ "github.com/go-sql-driver/mysql" // 引入 MySQL 驱动
)

func initMySQL() (*db.DB, error) {
    dsn := "root:123456@tcp(127.0.0.1:3306)/godeniter_demo?charset=utf8mb4&parseTime=True&loc=Local"
    database, err := db.Open("mysql", dsn)
    if err != nil {
        return nil, err
    }

    // 生产环境关键连接池调优
    database.SetMaxOpenConns(50)                 // 最大并发连接数 (根据业务流量调整)
    database.SetMaxIdleConns(10)                 // 空闲连接池保持数 (避免频繁创建/销毁连接)
    database.SetConnMaxLifetime(time.Hour)       // 连接最大生命周期 (防止 MySQL wait_timeout 断开)
    database.SetConnMaxIdleTime(10 * time.Minute) // 空闲连接最大超时

    // 测试连通性
    if err := database.Ping(); err != nil {
        return nil, err
    }

    return database, nil
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

---

## 5. MySQL 实战完整 Web 业务示例 (Full Runnable Demo)

以下展示如何在实际 Godeniter 工程中接入 MySQL，完成**表结构设计、数据模型映射、依赖注入及完整 RESTful API 增删改查**。

### (1) MySQL 数据表结构设计 (DDL)

```sql
CREATE DATABASE IF NOT EXISTS `godeniter_demo` DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
USE `godeniter_demo`;

CREATE TABLE IF NOT EXISTS `articles` (
    `id` INT AUTO_INCREMENT PRIMARY KEY,
    `title` VARCHAR(255) NOT NULL COMMENT '文章标题',
    `content` TEXT NOT NULL COMMENT '文章内容',
    `author` VARCHAR(100) NOT NULL DEFAULT '匿名' COMMENT '作者',
    `views` INT NOT NULL DEFAULT 0 COMMENT '浏览量',
    `status` TINYINT NOT NULL DEFAULT 1 COMMENT '状态: 1正常 0禁用',
    `created_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '发布时间',
    `updated_at` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='文章表';

-- 插入初始化测试数据
INSERT INTO `articles` (`title`, `content`, `author`, `views`, `status`) VALUES
('Godeniter 框架实战', '基于纯 Go 标准库与 MySQL 打造的高性能轻量 Web 服务。', 'admin', 10, 1),
('MySQL 索引与连接池调优', '生产环境中合理配置 MaxOpenConns 与 MaxIdleConns 能显著提升并发性能。', 'xbt', 88, 1);
```

### (2) 数据模型定义 (`models/article.go`)

```go
package models

import "time"

// Article 文章实体（通过 db 标签对应 MySQL 字段，json 标签对应接口输出）
type Article struct {
    ID        int       `db:"id"         json:"id"`
    Title     string    `db:"title"      json:"title"`
    Content   string    `db:"content"    json:"content"`
    Author    string    `db:"author"     json:"author"`
    Views     int       `db:"views"      json:"views"`
    Status    int       `db:"status"     json:"status"`
    CreatedAt time.Time `db:"created_at" json:"created_at"`
    UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}

// CreateArticleDTO 新增文章请求入参及表单验证
type CreateArticleDTO struct {
    Title   string `json:"title"   binding:"required,min=2,max=100"`
    Content string `json:"content" binding:"required,min=5"`
    Author  string `json:"author"  binding:"required"`
}
```

### (3) Web 服务启动与依赖注入 (`main.go`)

在主入口中初始化 MySQL 连接，通过 `app.Map(database)` 注入到依赖容器中，并在 Handler 中直接按需获取：

```go
package main

import (
    "fmt"
    "strconv"
    "time"

    "github.com/xbt/godeniter"
    "github.com/xbt/godeniter/db"
    _ "github.com/go-sql-driver/mysql"
)

func main() {
    // 1. 初始化 MySQL 连接池
    dsn := "root:123456@tcp(127.0.0.1:3306)/godeniter_demo?charset=utf8mb4&parseTime=True&loc=Local"
    database, err := db.Open("mysql", dsn)
    if err != nil {
        panic(fmt.Sprintf("MySQL 连接失败: %v", err))
    }
    database.SetMaxOpenConns(50)
    database.SetMaxIdleConns(10)
    database.SetConnMaxLifetime(time.Hour)

    if err := database.Ping(); err != nil {
        panic(fmt.Sprintf("MySQL Ping 失败: %v", err))
    }

    // 2. 初始化经典引擎并注入数据库实例
    app := godeniter.Classic()
    app.Map(database) // 全局依赖注入，所有 Handler 可直接声明 *db.DB

    // 3. 注册文章 CRUD 接口路由
    api := app.Group("/api/v1/articles")
    {
        // 1. 分页检索与模糊搜索
        api.Get("", func(c *godeniter.Context, database *db.DB) {
            keyword := c.Query("keyword")
            page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
            pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

            qb := database.Table("articles").Where("status = ?", 1)
            if keyword != "" {
                qb.Like("title", "%"+keyword+"%")
            }

            var list []Article
            pager, err := qb.OrderBy("id DESC").Paginate(page, pageSize, &list)
            if err != nil {
                c.Fail(500, "查询失败: "+err.Error())
                return
            }
            c.Success(godeniter.H{
                "list":  list,
                "pager": pager,
            })
        })

        // 2. 查看详情并自增阅读量
        api.Get("/:id", func(c *godeniter.Context, database *db.DB) {
            id := c.Param("id")
            var article Article
            err := database.Table("articles").Where("id = ?", id).First(&article)
            if err != nil {
                c.Fail(404, "文章不存在")
                return
            }

            // 阅读量自增
            _, _ = database.Table("articles").Where("id = ?", id).Increment("views", 1)
            article.Views++

            c.Success(article)
        })

        // 3. 新增文章 (带参数校验)
        api.Post("", func(c *godeniter.Context, database *db.DB) {
            var form CreateArticleDTO
            if err := c.BindAndValidate(&form); err != nil {
                c.Fail(400, "参数错误: "+err.Error())
                return
            }

            insertID, err := database.Table("articles").Insert(map[string]any{
                "title":      form.Title,
                "content":    form.Content,
                "author":     form.Author,
                "views":      0,
                "status":     1,
                "created_at": time.Now(),
                "updated_at": time.Now(),
            })
            if err != nil {
                c.Fail(500, "保存失败: "+err.Error())
                return
            }

            c.Success(godeniter.H{"id": insertID, "message": "发布成功"})
        })

        // 4. 更新文章
        api.Put("/:id", func(c *godeniter.Context, database *db.DB) {
            id := c.Param("id")
            title := c.PostForm("title")
            content := c.PostForm("content")

            updateData := map[string]any{"updated_at": time.Now()}
            if title != "" {
                updateData["title"] = title
            }
            if content != "" {
                updateData["content"] = content
            }

            rows, err := database.Table("articles").Where("id = ?", id).Update(updateData)
            if err != nil || rows == 0 {
                c.Fail(500, "更新失败或未修改")
                return
            }
            c.Success("更新成功")
        })

        // 5. 删除文章 (逻辑软删除示例)
        api.Delete("/:id", func(c *godeniter.Context, database *db.DB) {
            id := c.Param("id")
            // 软删除: 将 status 置为 0
            rows, err := database.Table("articles").Where("id = ?", id).Update(map[string]any{"status": 0})
            if err != nil || rows == 0 {
                c.Fail(500, "删除失败")
                return
            }
            c.Success("删除成功")
        })

        // 6. 事务操作示例 (如批量将某作者的文章归档)
        api.Post("/archive-author", func(c *godeniter.Context, database *db.DB) {
            author := c.Query("author")
            err := database.Transaction(func(tx *db.Tx) error {
                if _, err := tx.Table("articles").Where("author = ?", author).Update(map[string]any{"status": 0}); err != nil {
                    return err // 遇到错误自动 Rollback
                }
                return nil // 正常结束自动 Commit
            })
            if err != nil {
                c.Fail(500, "事务执行失败: "+err.Error())
                return
            }
            c.Success("作者文章已全部安全归档")
        })
    }

    // 4. 启动服务
    _ = app.Run(":8080")
}
```

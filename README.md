# Godeniter 2.0 Web 开发框架

<p align="center">
  <b>极简 · 零外部依赖 · 依赖注入 · 经典 MVC · 单文件极速分发</b>
</p>

---

## 📖 框架简介

**Godeniter** 是一套专为快速开发、极简部署而设计的 Go 语言 Web 框架。
设计灵感汲取自 **PHP 经典框架 (CodeIgniter / Laravel)** 的开发直觉与 **Martini** 优雅的依赖注入哲学：

* **100% 纯 Go 标准库（0 外部依赖）**：核心模块完全基于 Go 1.20+ 原生库开发，体积轻巧，无任何第三方包供应链风险。
* **Martini 风格依赖注入 (Dependency Injection)**：Handler 签名完全自由，按需声明参数，框架在运行时通过轻量反射容器自动注入。
* **高性能 Trie 树动态路由**：支持 RESTful 动词、`:param` 动态命名参数、`*filepath` 全路径通配符、路由多级分组与分组中间件。
* **洋葱圈中间件栈**：提供标准的 `c.Next()` 与 `c.Abort()` 流水线机制，内置 Logger、Recovery 与 CORS 中间件。
* **PHP 风格轻量 QueryBuilder**：基于标准库 `database/sql` 封装链式 SQL 构建器，支持 SQLite 与 MySQL，提供直观的 ActiveRecord 体验。
* **单二进制打包与 Windows 客户机交付**：通过 `go:embed` 将前端模板与静态资源内嵌，编译生成的单文件 `.exe` 在 Windows 客户机上双击即可直接运行，并在控制台实时输出访问地址。

---

## 🚀 快速上手 (Quick Start)

### 1. 基础示例

```go
package main

import (
    "godeniter"
    "godeniter/router"
    "net/http"
)

func main() {
    // 1. 创建带有 Logger 和 Recovery 的经典引擎
    app := godeniter.Classic()

    // 2. 标准 Context 路由
    app.Get("/ping", func(c *godeniter.Context) {
        c.JSON(http.StatusOK, godeniter.H{"message": "pong"})
    })

    // 3. 动态路由参数
    app.Get("/users/:id", func(params router.Params) (int, godeniter.H) {
        return 200, godeniter.H{"user_id": params.Get("id")}
    })

    // 4. 路由分组
    api := app.Group("/api/v1")
    {
        api.Get("/version", func() string {
            return "v2.0.0"
        })
    }

    // 5. 启动服务 (终端自动打印访问链接)
    app.Run(":8080")
}
```

---

## 💉 依赖注入系统 (Dependency Injection)

Godeniter 内置了基于反射实现的轻量依赖注入容器，支持将服务（如配置、数据库连接、用户上下文）映射到容器中：

```go
type Database struct { /* ... */ }
type AppConfig struct { Env string }

func main() {
    app := godeniter.New()

    // 映射全局依赖
    app.Map(&AppConfig{Env: "production"})
    app.Map(&Database{})

    // Handler 任意声明所需依赖，框架自动注入！
    app.Get("/status", func(cfg *AppConfig, db *Database) (int, string) {
        return 200, "App Env: " + cfg.Env
    })
}
```

---

## 🗄️ 数据库操作 (QueryBuilder)

基于 Go 标准库 `database/sql`，支持类似 CodeIgniter 3 的链式调用：

```go
type User struct {
    ID       int    `db:"id"`
    Username string `db:"username"`
    Age      int    `db:"age"`
}

// 1. 查询多条
var users []User
err := db.Table("users").
    Select("id", "username", "age").
    Where("age >= ?", 18).
    OrderBy("id DESC").
    Limit(10).
    Find(&users)

// 2. 查询单条
var user User
err := db.Table("users").Where("id = ?", 1).First(&user)

// 3. 插入记录 (支持 map 或 struct)
id, err := db.Table("users").Insert(map[string]any{
    "username": "ben",
    "age":      28,
})

// 4. 更新记录
rows, err := db.Table("users").Where("id = ?", 1).Update(map[string]any{
    "age": 29,
})

// 5. 事务处理
err := db.Transaction(func(tx *db.Tx) error {
    _, err := tx.Table("accounts").Where("id = ?", 1).Update(...)
    return err
})
```

> **数据库驱动引入说明**：  
> 框架底层为纯标准库 0 依赖，在使用具体数据库时只需在入口匿名导入标准驱动：
> * **MySQL**: `import _ "github.com/go-sql-driver/mysql"`
> * **SQLite**: `import _ "modernc.org/sqlite"`（纯 Go 实现，无 CGO 依赖，完美支持跨平台交叉编译）。

---

## 📦 单二进制打包与 Windows 客户机交付

利用 Go 1.16+ 原生 `embed` 特性，可以将前端页面和静态资源全部打入一个可执行二进制文件中：

```go
package main

import (
    "embed"
    "godeniter"
    "html/template"
    "io/fs"
)

//go:embed views/*
var viewsFS embed.FS

func main() {
    app := godeniter.Classic()

    // 加载嵌入式模板
    subViews, _ := fs.Sub(viewsFS, "views")
    app.SetHTMLTemplate(template.Must(template.ParseFS(subViews, "*.html")))

    app.Get("/", func(c *godeniter.Context) {
        c.HTML(200, "index.html", godeniter.H{"Title": "欢迎使用 Godeniter"})
    })

    app.Run(":8080")
}
```

### 一键构建命令

* **macOS / Linux 环境下生成 Windows `.exe`**：
  ```bash
  CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o godeniter-app.exe ./cmd/server
  ```
* **运行构建脚本**：
  ```bash
  ./build.sh      # macOS / Linux
  build.bat       # Windows
  ```

构建生成的 `godeniter-app.exe` 无需安装任何环境，直接拷贝给客户，**双击即可运行**并在终端提示访问地址（如 `http://127.0.0.1:8080`）。

---

## 📂 项目模块结构

* [`inject/`](file:///Users/ben/godeniter/inject/) - 依赖注入容器核心（`Map`, `MapTo`, `Invoke`, `Apply`）
* [`router/`](file:///Users/ben/godeniter/router/) - 前缀树（Trie）路由器、路由分组与参数解析
* [`context.go`](file:///Users/ben/godeniter/context.go) - 请求上下文、洋葱圈流转与多格式渲染
* [`godeniter.go`](file:///Users/ben/godeniter/godeniter.go) - 核心引擎入口、模板加载与服务启动
* [`middleware/`](file:///Users/ben/godeniter/middleware/) - 内置中间件（Logger、Recovery、CORS）
* [`db/`](file:///Users/ben/godeniter/db/) - 数据库连接管理与链式 QueryBuilder
* [`cmd/server/`](file:///Users/ben/godeniter/cmd/server/) - 完整可独立打包交付的演示应用

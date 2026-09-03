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
## 📦 如何在你的新项目中引用 Godeniter

Godeniter 本身是一个标准且纯粹的 Go Module，你可以直接在任何新项目中引用它：

```bash
# 在新项目目录下执行初始化并引用
go mod init my-app
# 如果是本地项目或未发布到远程，可以使用 replace 或直接引用
# 比如本地开发：go mod edit -replace godeniter=/path/to/godeniter
```

在代码中直接导入：
```go
import (
    "godeniter"
    "godeniter/router"
    "godeniter/middleware"
    "godeniter/db"
)
```

---

## 🎯 官方内置全功能开箱即用 Demo

为了满足不同开发场景的需求，框架在 `examples/` 目录下内置了两种最具代表性的完整项目解决方案，均支持 **单文件嵌入与一键打包为 Windows `.exe`**：

### 1. [模式一：前后端分离 (RESTful API + SPA 单页) 方案](file:///Users/ben/godeniter/examples/01_api_spa/)
* **适用场景**：现代化管理后台、移动端/小程序后端、Vue/React 前后端分离项目。
* **包含特性**：
  * 统一 API 响应格式封装：`c.Success(data)` / `c.Fail(code, msg)`
  * 跨域支持：内置 `middleware.CORS()`
  * 控制器依赖注入：自动将 `ArticleService` 注入到 Handler 中
  * 内嵌 SPA 前端单页界面，一键编译为 `.exe`，双击即用。
* **快速运行**：`go run ./examples/01_api_spa/main.go`

### 2. [模式二：经典 PHP 风格服务端渲染 (MVC + HTML Template) 方案](file:///Users/ben/godeniter/examples/02_mvc_template/)
* **适用场景**：企业内部进销存 ERP、WMS 客户端、SEO 友好的官网或内容管理系统。
* **包含特性**：
  * 经典的 MVC 目录分层（`controllers/`, `models/`, `views/`）
  * 原生 `html/template` 服务端模板数据循环与条件分支渲染
  * 传统 HTML 表单 POST 提交验证与页面重定向（`c.Redirect`）
  * Session / Cookie 登录状态管理（`c.SetCookie` / `c.Cookie`）
* **快速运行**：`go run ./examples/02_mvc_template/main.go`

---

## 🚀 基础快速上手 (Quick Start)

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
  # 编译 Demo 1 (API + SPA 模式)
  CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o dist/demo1_api_spa.exe ./examples/01_api_spa/main.go

  # 编译 Demo 2 (MVC 模板渲染模式)
  CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o dist/demo2_mvc_template.exe ./examples/02_mvc_template/main.go
  ```
* **运行一键构建脚本**：
  ```bash
  ./build.sh      # macOS / Linux (一键编译所有示例并生成 dist/*.exe)
  build.bat       # Windows (一键编译所有示例)
  ```

构建生成的 `.exe` 文件无需安装任何环境，直接拷贝给客户，**双击即可运行**并在终端提示访问地址（如 `http://127.0.0.1:8080`）。

---

## 📂 项目模块结构

* [`inject/`](file:///Users/ben/godeniter/inject/) - 依赖注入容器核心（`Map`, `MapTo`, `Invoke`, `Apply`）
* [`router/`](file:///Users/ben/godeniter/router/) - 前缀树（Trie）路由器、路由分组与参数解析
* [`context.go`](file:///Users/ben/godeniter/context.go) - 请求上下文、洋葱圈流转与多格式渲染（JSON/HTML/Data）
* [`godeniter.go`](file:///Users/ben/godeniter/godeniter.go) - 核心引擎入口、模板加载与服务启动
* [`middleware/`](file:///Users/ben/godeniter/middleware/) - 内置中间件（Logger、Recovery、CORS）
* [`db/`](file:///Users/ben/godeniter/db/) - 数据库连接管理与链式 QueryBuilder
* [`examples/01_api_spa/`](file:///Users/ben/godeniter/examples/01_api_spa/) - 模式一：前后端分离 RESTful API + SPA 单页完整 Demo
* [`examples/02_mvc_template/`](file:///Users/ben/godeniter/examples/02_mvc_template/) - 模式二：经典 PHP 风格服务端渲染 MVC + HTML Template 完整 Demo
* [`dist/`](file:///Users/ben/godeniter/dist/) - 编译生成的跨平台单文件可执行程序输出目录

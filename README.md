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
## 📦 如何在项目中引用 Godeniter

### 方式一：直接通过标准 `go get` 安装 ⭐ (最简便)

```bash
# 在您的新项目目录下
go mod init my-app
go get -u github.com/xbt/godeniter
```

在代码中直接导入使用：
```go
import (
    "github.com/xbt/godeniter"
    "github.com/xbt/godeniter/router"
    "github.com/xbt/godeniter/middleware"
    "github.com/xbt/godeniter/db"
    "github.com/xbt/godeniter/utils/str"
    "github.com/xbt/godeniter/utils/upload"
)
```

### 方式二：使用官方开箱即用脚手架项目 ([`godeniter-starter`](https://github.com/xbt/godeniter-starter))

我们提供了预先搭建好标准分层结构（配置、控制器、模型、业务层、中间件、视图与一键打包脚本）的独立模板工程：

```bash
# 进入 starter 工程
cd ../godeniter-starter

# 极速本地启动
go run main.go
# 浏览器访问: http://127.0.0.1:8080 (带管理后台、Session 登录与 API 演示)

# 一键编译为 Windows 独立单文件 .exe
./build.sh
```

---

## 📴 离线与受限网络开发支持 (Zip 包即用)

Godeniter 采用 **100% 纯 Go 标准库（0 外部第三方依赖）** 设计。在**内网断网或受限网络**环境下，无需拉取任何外部 Go 模块：
* **Zip 源码包即用**：解压即可直接 `go run` 运行 Demo。
* **本地路径引用 (`replace`)**：支持在 `go.mod` 中通过 `replace` 直接读取本地源码。
* **终极离线交付**：一键打包为 Windows 单文件 `.exe`，免安装 Go 环境与网络，双击直接运行。

👉 详见专有手册：[**《离线环境与受限网络开发指南 (docs/offline.md)》**](./docs/offline.md)

---

## ⚡ 极速体验内置 Demo (一键启动运行)

框架在 `examples/` 目录下内置了两种开箱即用的完整项目模式，克隆仓库后可直接在终端一键运行体验：

```bash
# 1. 运行模式一：前后端分离 (RESTful API + SPA 单页 + 文件上传 + 分页搜索)
go run ./examples/01_api_spa/main.go
# 浏览器直接打开访问: http://127.0.0.1:8080 (内置交互式控制台与 CodeIgniter 开发者手册)

# 2. 运行模式二：经典 PHP 风格服务端渲染 (MVC + HTML Template + Session 登录)
go run ./examples/02_mvc_template/main.go
# 浏览器直接打开访问: http://127.0.0.1:8080 (默认账号: admin / 密码: 123456)
```

---

## 🎯 官方内置全功能开箱即用 Demo 详述

### 1. [模式一：前后端分离 (RESTful API + SPA 单页) 方案](./examples/01_api_spa/)
* **适用场景**：现代化管理后台、移动端/小程序后端、Vue/React 前后端分离项目。
* **包含特性**：
  * 统一 API 响应格式封装：`c.Success(data)` / `c.Fail(code, msg)`
  * 跨域支持：内置 `middleware.CORS()`
  * 控制器依赖注入：自动将 `ArticleService` 注入到 Handler 中
  * 封面图片异步上传与预览、文章分页检索与标题关键词模糊搜索
  * 内嵌现代化管理面板与交互式开发者手册，一键编译为 `.exe`，双击即用。
* **快速运行命令**：
  ```bash
  go run ./examples/01_api_spa/main.go
  ```

### 2. [模式二：经典 PHP 风格服务端渲染 (MVC + HTML Template) 方案](./examples/02_mvc_template/)
* **适用场景**：企业内部进销存 ERP、WMS 客户端、SEO 友好的官网或内容管理系统。
* **包含特性**：
  * 经典的 MVC 目录分层（`controllers/`, `models/`, `views/`）
  * 原生 `html/template` 服务端模板数据循环与条件分支渲染
  * 传统 HTML 表单 POST 提交验证与页面重定向（`c.Redirect`）
  * Session / Cookie 登录状态管理（`c.SetCookie` / `c.Cookie`）与头像上传
* **快速运行命令**：
  ```bash
  go run ./examples/02_mvc_template/main.go
  ```

---

## 🚀 基础快速上手 (Quick Start)

### 1. 基础示例

```go
package main

import (
    "github.com/xbt/godeniter"
    "github.com/xbt/godeniter/router"
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

---

## 🗄️ 数据库操作与 ActiveRecord (类似 CodeIgniter 3 增强版)

基于 Go 标准库 `database/sql`，提供极度舒适的链式构造器与一键分页能力：

```go
type User struct {
    ID       int    `db:"id"`
    Username string `db:"username"`
    Email    string `db:"email"`
    Age      int    `db:"age"`
    Avatar   string `db:"avatar"`
}

// 1. 模糊搜索 (Like) 与范围过滤 (Between)
var users []User
err := db.Table("users").
    Select("users.id", "users.username", "profiles.avatar").
    LeftJoin("profiles", "users.id = profiles.user_id").
    Where("status = ?", 1).
    Like("username", "%admin%").
    WhereBetween("age", 18, 60).
    OrderBy("users.id DESC").
    Find(&users)

// 2. 一键分页查询 (Paginate)
var list []User
pager, err := db.Table("users").
    Where("status = ?", 1).
    OrderBy("id DESC").
    Paginate(1, 10, &list)
// pager 包含: Total (总数), TotalPages (总页数), Page (当前页), PageSize, HasNext, HasPrev

// 3. 聚合统计与快捷自增 (Increment / Sum / Avg / Count)
count, _ := db.Table("users").Count()
sumViews, _ := db.Table("articles").Sum("views")
// 快捷自增阅读量
rows, _ := db.Table("articles").Where("id = ?", 1).Increment("views", 1)

// 4. 批量插入 (InsertBatch)
records := []any{
    map[string]any{"username": "user1", "age": 20},
    map[string]any{"username": "user2", "age": 22},
}
_, err = db.Table("users").InsertBatch(records)

// 5. 事务保护 (自动 Commit / Panic / Error 回滚)
err := db.Transaction(func(tx *db.Tx) error {
    _, err := tx.Table("accounts").Where("id = ?", 1).Update(map[string]any{"balance": 500})
    return err
})
```

---

## 📁 文件上传与安全存储 (`utils/upload`)

提供对标 CodeIgniter Upload 类的一行代码上传与安全校验：

```go
import "github.com/xbt/godeniter/utils/upload"

app.Post("/upload/avatar", func(c *godeniter.Context) {
    file, err := c.FormFile("avatar")
    if err != nil {
        c.Fail(400, "请选择文件")
        return
    }

    opts := upload.Options{
        SaveDir:     "./uploads/avatars",
        MaxBytes:    2 * 1024 * 1024, // 限制 2MB
        AllowedExts: []string{".jpg", ".png", ".jpeg"},
        AutoRename:  true, // 自动重命名为 20260903_120000_random.jpg
    }

    savedPath, err := c.SaveUploadedFileWithOptions(file, opts)
    if err != nil {
        c.Fail(500, err.Error())
        return
    }

    c.Success(godeniter.H{"url": "/" + savedPath})
})
```

---

## 🔤 字符串与安全辅助工具库 (`utils/str`)

对标 CodeIgniter `string_helper`, `text_helper`, `security_helper`：

```go
import "github.com/xbt/godeniter/utils/str"

// 1. 随机生成
code := str.RandomNumeric(6)     // 6位数字验证码 (如 "839201")
token := str.Random(32)          // 32位安全 Token
uuid := str.UUID()               // 标准 UUIDv4

// 2. 命名互转
snake := str.SnakeCase("UserProfile")    // "user_profile"
camel := str.CamelCase("user_profile", false) // "userProfile"
kebab := str.KebabCase("UserProfile")    // "user-profile"

// 3. 文本处理与中文截断
summary := str.Truncate("这是一篇非常长的文章正文...", 10, "...")
sub := str.Substr("你好Godeniter世界", 2, 9) // "Godeniter"

// 4. 敏感信息脱敏
phone := str.MaskPhone("13812345678") // "138****5678"
email := str.MaskEmail("user@godeniter.dev") // "u***r@godeniter.dev"
idCard := str.MaskIDCard("110101199003072345") // "1101**********2345"

// 5. 哈希与 XSS 安全过滤
md5Val := str.MD5("123456")
safeHTML := str.XSSFilter("<script>alert(1)</script>")
```

---

## 📦 单二进制打包与 Windows 客户机交付

利用 Go 1.16+ 原生 `embed` 特性，可以将前端页面和静态资源全部打入一个可执行二进制文件中：

```go
package main

import (
    "embed"
    "github.com/xbt/godeniter"
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

## 🛡️ 参数绑定与轻量验证器 (`binding/`)

Godeniter 内置了基于 Struct Tag 的纯标准库数据验证引擎（支持 `required`, `min=N`, `max=N`, `email`, `numeric` 规则）：

```go
type RegisterForm struct {
    Username string `json:"username" form:"username" binding:"required,min=4,max=16"`
    Email    string `json:"email"    form:"email"    binding:"required,email"`
    Age      int    `json:"age"      form:"age"      binding:"required,min=18"`
}

app.Post("/register", func(c *godeniter.Context) {
    var form RegisterForm
    // 自动根据 Content-Type 解析并执行 struct tag 规则校验
    if err := c.BindAndValidate(&form); err != nil {
        c.Fail(40001, "表单验证失败: " + err.Error())
        return
    }
    c.Success(godeniter.H{"user": form.Username})
})
```

---

## 🔐 服务端会话管理 (`session/`)

提供类似 PHP `$_SESSION` 的极简体验，基于 HMAC-SHA256 安全签名的防篡改 CookieStore，支持开箱即用的无状态单文件交付：

```go
import "github.com/xbt/godeniter/session"

func main() {
    app := godeniter.Classic()

    // 挂载 Session 中间件
    store := session.NewCookieStore("my-secret-key-123456")
    app.Use(godeniter.Session(store, "app_session"))

    // 在 Controller 中直接注入 session.Session 使用
    app.Post("/login", func(c *godeniter.Context, sess session.Session) {
        sess.Set("user_id", 1001)
        sess.Set("username", "ben")
        c.Redirect(302, "/dashboard")
    })

    app.Get("/dashboard", func(c *godeniter.Context, sess session.Session) {
        username := sess.GetString("username")
        c.String(200, "欢迎回来: " + username)
    })
}
```

---

## 🛠️ 官方 CLI 脚手架工具 (`cmd/godeniter`)

类似于 `php artisan` 或 `codeigniter spark`，支持命令行一键初始化全新独立工程：

```bash
# 查看版本与帮助
godeniter version
godeniter help

# 一键创建前后端分离 API 模版工程
godeniter new my-api-project --template=api

# 一键创建经典 PHP 风格 MVC 模板渲染工程
godeniter new my-web-project --template=mvc
```

---

## 📂 项目模块结构

* **根目录核心模块**（极简纯粹，仅保留启动与核心上下文）：
  * [`godeniter.go`](./godeniter.go) - 核心引擎入口、模板加载与服务启动
  * [`context.go`](./context.go) - 请求上下文、洋葱圈流转与多格式渲染（JSON/HTML/Data）
  * [`response_writer.go`](./response_writer.go) - 状态码与响应体双重拦截包装器
* **专业功能子模块**：
  * [`inject/`](./inject/) - 依赖注入容器核心（`Map`, `MapTo`, `Invoke`, `Apply`）
  * [`router/`](./router/) - 前缀树（Trie）路由器、路由分组与参数解析
  * [`db/`](./db/) - 数据库连接管理与增强型 ActiveRecord QueryBuilder (Like, Join, Paginate)
  * [`session/`](./session/) - 服务端会话管理与安全签名 CookieStore
  * [`binding/`](./binding/) - 请求参数绑定与基于 Struct Tag 的轻量验证器
  * [`middleware/`](./middleware/) - 内置中间件（Logger、Recovery、CORS）
* **通用工具集 (`utils/`)**：
  * [`utils/str/`](./utils/str/) - 字符串处理、脱敏、哈希与随机生成
  * [`utils/upload/`](./utils/upload/) - 文件上传安全处理类与存储校验器
* **命令行与开箱即用示例**：
  * [`cmd/godeniter/`](./cmd/godeniter/) - 官方 CLI 脚手架工具 (`godeniter new`)
  * [`examples/01_api_spa/`](./examples/01_api_spa/) - 模式一：前后端分离 RESTful API + SPA 单页 (带文件上传与分页) 完整 Demo
  * [`examples/02_mvc_template/`](./examples/02_mvc_template/) - 模式二：经典 PHP 风格服务端渲染 MVC + HTML Template (带头像上传与搜索) 完整 Demo
  * [`dist/`](./dist/) - 编译生成的跨平台单文件可执行程序输出目录

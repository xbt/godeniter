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
```

---

## 📴 离线与受限网络开发支持 (Zip 包即用)

Godeniter 采用 **100% 纯 Go 标准库（0 外部第三方依赖）** 设计。在**内网断网或受限网络**环境下，无需拉取任何外部 Go 模块：
* **Zip 源码包即用**：解压即可直接 `go run` 运行 Demo。
* **本地路径引用 (`replace`)**：支持在 `go.mod` 中通过 `replace` 直接读取本地源码。
* **终极离线交付**：一键打包为 Windows 单文件 `.exe`，免安装 Go 环境与网络，双击直接运行。

👉 详见专有手册：[**《离线环境与受限网络开发指南 (docs/offline.md)》**](./docs/offline.md)

---

## ⚙️ 动态配置与外部配置文件 (0 依赖 JSON)

Godeniter 坚持 **0 外部第三方库依赖**，推荐并内置基于 Go 原生 `encoding/json` 的分层动态配置体系（`config.json`）：
* **Sidecar 伴随自生成**：首次打包为 `.exe` 或部署运行若无配置文件，自动在程序就近生成结构清晰、带注释字段的 `config.json` 模板。
* **三层动态覆盖**：代码硬编码默认值 -> 本地 `config.json` -> 生产环境环境变量（`APP_PORT`, `DB_DSN` 等），优先级依次递增。
* **数据库连接灵活切换**：开箱即用支持 SQLite3 单文件数据库与外部 MySQL/PostgreSQL 生产库平滑切换。
* **客户机零门槛改配置**：交付给客户的单文件 `.exe` 无需重新编译，客户用记事本打开 `config.json` 即可随意更改端口与数据库。

👉 详见专有手册：[**《动态配置与数据库连接手册 (docs/config.md)》**](./docs/config.md)

---

## 🛡️ 服务生命周期与守护进程管理 (Daemon & Lifecycle)

Godeniter 原生内置基于纯 Go 标准库的跨平台服务生命周期与守护进程管理器（`godeniter/daemon`）：
* **双模自由切换**：默认无参数时为**前台开发调试模式**（输出彩色 Banner，随时 `Ctrl+C` 退出）；输入 `start` 或开启配置则自动进入**后台守护模式**（Setsid 自动脱离终端，输出 PID 并**立即返回命令行**，断开 SSH / 关闭终端持续运行）。
* **类似 Nginx 的常用指令集**：
  * `go run main.go start`（或 `./app start`）：后台静默启动服务，重定向日志至 `app.log`；
  * `go run main.go status`（或 `./app status`）：查看当前服务运行状态、PID 与监听端口；
  * `go run main.go stop`（或 `./app stop`）：向后台进程发送信号触发平滑优雅停机并清理 PID 文件；
  * `go run main.go restart`（或 `./app restart`）：平滑重启服务。
* **100% 跨平台兼容**：在 Linux/macOS 平台使用系统调用脱离终端会话；在 Windows 平台使用 `DETACHED_PROCESS` 剥离窗口，交叉编译零报错。

👉 详见专有手册：[**《服务生命周期管理与守护进程运维手册 (docs/daemon.md)》**](./docs/daemon.md)

---

## 🚀 跨平台系统托盘与桌面状态栏 (`tray/`)

Godeniter 原生内置跨平台桌面系统托盘与状态栏常驻客户端模式（`godeniter/tray`），满足将 Web 服务分发为独立桌面客户端的场景：
* **跨平台原生桌面交互**：
  * **macOS**：在屏幕右上角菜单栏常驻小图标，点击展开原生下拉菜单；
  * **Windows**：在屏幕右下角通知区域常驻托盘图标，支持右键原生菜单与双击快捷打开；
  * **Linux / 无头环境**：自动平滑降级为命令行信号监听模式。
* **开箱即用经典菜单体系**：
  * 🌐 **打开管理后台**：自动调起系统默认浏览器访问 Web 页面；
  * 📁 **打开应用目录**：直接调起系统文件管理器（Windows 资源管理器 / Mac Finder）定位到应用目录，方便用户寻找 `config.json`、`app.log` 或 SQLite 数据文件；
  * ℹ️ **关于系统**：弹窗展示当前系统版本、进程 PID、端口与运行环境；
  * ──────────────────（原生视觉分割线）
  * ⏹️ **退出程序**：平滑停机并安全退出进程；
  * 🧩 **自由扩展**：可通过 `Menus` 传入任意数量自定义功能菜单项。
* **贯彻 0 第三方依赖理念**：
  * Windows 端使用 **100% 纯 Go 标准库 `syscall`** 动态链接 Win32 原生 API，**无任何 CGO 依赖**，Mac 交叉编译 Windows `.exe` 畅通无阻！
  * macOS 端直接对接系统出厂内置的 Cocoa 框架，无需安装任何外部环境。

### 托盘模式快速调用示例：

```go
package main

import (
    "github.com/xbt/godeniter"
    "github.com/xbt/godeniter/tray"
)

func main() {
    app := godeniter.Classic()
    app.Get("/", func(c *godeniter.Context) {
        c.String(200, "Hello Godeniter Desktop!")
    })

    // 在后台协程启动 Web 监听
    go app.Run(":8080")

    // 在主线程启动跨平台原生系统托盘
    _ = tray.Run(tray.Options{
        Title:   "Godeniter Desktop",
        Tooltip: "Godeniter 桌面客户端",
        URL:     "http://127.0.0.1:8080",
        Version: "v1.0.0",
        Port:    ":8080",
    })
}
```

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
  * 内嵌现代化管理面板与交互式开发者手册。
* **快速运行命令**：
  ```bash
  go run ./examples/01_api_spa/main.go
  ```

### 2. [模式二：经典 PHP 风格服务端渲染 (MVC + HTML Template 套页面) 方案](./examples/02_mvc_template/)
* **适用场景**：企业内部进销存 ERP、WMS 客户端、SEO 友好的官网或内容管理系统。
* **包含特性**：
  * 经典的 MVC 目录分层（`controllers/`, `models/`, `views/`）
  * 原生 `html/template` 服务端模板数据循环与条件分支渲染
  * 传统 HTML 表单 POST 提交验证与页面重定向（`c.Redirect`）
  * Session / Cookie 登录状态管理（`c.SetCookie` / `c.Cookie`）与头像上传。
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

    // 3. 动态路由参数与 Query 默认值解析
    app.Get("/users/:id", func(c *godeniter.Context) {
        userID := c.Param("id")
        page := c.QueryInt("page", 1) // 获取 ?page=2，未传或非法自动取默认值 1
        c.JSON(200, godeniter.H{"user_id": userID, "page": page})
    })

    // 4. 路由分组
    api := app.Group("/api/v1")
    {
        api.Get("/version", func() string {
            return "v2.0.0"
        })
    }

    // 5. 启动服务 (终端自动打印访问链接，内置平滑优雅停机 Graceful Shutdown)
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

## 🗄️ 数据库操作与 ActiveRecord (支持 MySQL 与 SQLite)

底层基于 Go 原生 `database/sql`，提供极度舒适的链式构造器、一键分页与事务能力。

### 1. 初始化数据库连接与连接池 (以 MySQL 为例)
```go
package main

import (
    "github.com/xbt/godeniter"
    "github.com/xbt/godeniter/db"
    _ "github.com/go-sql-driver/mysql" // 引入 MySQL 驱动 (或 _ "modernc.org/sqlite")
)

func main() {
    app := godeniter.Classic()

    // 1. 初始化 MySQL 连接池
    dsn := "root:123456@tcp(127.0.0.1:3306)/godeniter_demo?charset=utf8mb4&parseTime=True&loc=Local"
    database, err := db.Open("mysql", dsn)
    if err != nil {
        panic("数据库连接失败: " + err.Error())
    }
    database.SetMaxOpenConns(50) // 最大打开连接数
    database.SetMaxIdleConns(10) // 最大空闲连接数

    // 2. 注入全局依赖容器，所有 Controller/Handler 均可直接声明注入 *db.DB
    app.Map(database)

    // 3. 在 Handler 中直接使用 database 执行查询
    app.Get("/api/users", func(c *godeniter.Context, database *db.DB) {
        var users []User
        pager, err := database.Table("users").
            Where("status = ?", 1).
            Like("username", "%admin%").
            OrderBy("id DESC").
            Paginate(1, 10, &users)
        if err != nil {
            c.Fail(500, err.Error())
            return
        }
        c.Success(godeniter.H{"list": users, "pager": pager})
    })

    app.Run(":8080")
}
```

### 2. 丰富查询操作 (对比 CodeIgniter 3)

```go
type User struct {
    ID       int    `db:"id"       json:"id"`
    Username string `db:"username" json:"username"`
    Email    string `db:"email"    json:"email"`
    Age      int    `db:"age"      json:"age"`
    Avatar   string `db:"avatar"   json:"avatar"`
}

// 1. 模糊搜索 (Like) 与范围过滤 (Between)
var users []User
err := database.Table("users").
    Select("users.id", "users.username", "profiles.avatar").
    LeftJoin("profiles", "users.id = profiles.user_id").
    Where("status = ?", 1).
    Like("username", "%admin%").
    WhereBetween("age", 18, 60).
    OrderBy("users.id DESC").
    Find(&users)

// 2. 一键分页查询 (Paginate)
var list []User
pager, err := database.Table("users").
    Where("status = ?", 1).
    OrderBy("id DESC").
    Paginate(1, 10, &list)
// pager 包含: Total (总数), TotalPages (总页数), Page (当前页), PageSize, HasNext, HasPrev

// 3. 聚合统计与快捷自增 (Increment / Sum / Avg / Count)
count, _ := database.Table("users").Count()
sumViews, _ := database.Table("articles").Sum("views")
rows, _ := database.Table("articles").Where("id = ?", 1).Increment("views", 1)

// 4. 插入与批量插入 (Insert / InsertBatch)
id, _ := database.Table("users").Insert(map[string]any{"username": "ben", "age": 28})
records := []any{
    map[string]any{"username": "user1", "age": 20},
    map[string]any{"username": "user2", "age": 22},
}
_, err = database.Table("users").InsertBatch(records)

// 5. 事务保护 (自动 Commit / Panic / Error 回滚)
err := database.Transaction(func(tx *db.Tx) error {
    _, err := tx.Table("accounts").Where("id = ?", 1).Update(map[string]any{"balance": 500})
    return err
})
```

👉 详见专有手册：[**《数据库与 ActiveRecord 开发手册 (docs/database.md)》**](./docs/database.md) 与 [**MySQL 实战演示工程 (`examples/database_mysql/`)**](./examples/database_mysql/)（含完整 MySQL 表设计 DDL、全套 CRUD、分页、事务与交互式控制台）

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

## 🎨 服务端 HTML 模板渲染与无侵入注释语法 (Zero-Distortion Templates)

Godeniter 支持 Go 原生 `html/template`，并内置开箱即用的 `app.LoadHTMLFS(...)` 与 `app.LoadHTMLGlob(...)` 辅助函数。更进一步，Godeniter 独创支持**无侵入 HTML 注释模板语法**（`<!--{{ ... }}-->`）：

### 💡 为什么需要无侵入注释语法？
* **前端原型双击可预览**：传统的 `{{ if .User }}...{{ end }}` 在前端 UI 设计师或本地浏览器直接双击打开 `.html` 时，会被视为裸文本渲染，破坏 CSS 样式与 DOM 排版。
* **浏览器天然友好**：写成 `<!--{{ if .User }}-->...<!--{{ end }}-->` 时，静态浏览器会天然将其当成标准 HTML 注释隐藏，页面完全所见即所得！
* **100% 混用兼容**：标准 `{{ ... }}` 与 `<!--{{ ... }}-->` 可在同文件任意混用，框架在启动时进行预处理转译，运行时走 Go 原生 AST 编译，**零运行时损耗**。
* **零误伤**：普通的 HTML 注释（如 `<!-- 开发说明注释 -->`）完全保持原样，绝不误处理。

### 示例代码

```go
package main

import (
    "embed"
    "io/fs"
    "github.com/xbt/godeniter"
)

//go:embed views/*
var viewsFS embed.FS

func main() {
    app := godeniter.Classic()

    // 1. 一行代码加载嵌入式视图模板（自动支持 <!--{{ ... }}--> 与 {{ ... }}）
    subViews, _ := fs.Sub(viewsFS, "views")
    app.LoadHTMLFS(subViews, "*.html")

    app.Get("/", func(c *godeniter.Context) {
        c.HTML(200, "index.html", godeniter.H{
            "Title": "欢迎使用 Godeniter",
            "User":  "Admin",
        })
    })

    app.Run(":8080")
}
```

**模板文件 `views/detail.html` 示例：**
```html
<div class="user-card">
    <!-- 普通注释被原样保留 -->
    <!-- 下面使用无侵入注释语法，双击 HTML 原型时该编辑按钮在未登录时不乱码破坏布局 -->
    <!--{{ if .CurrentUser }}-->
        <a href="/admin/articles/edit/<!--{{ .Article.ID }}-->" class="edit-btn">✏️ 编辑文章</a>
    <!--{{ end }}-->
</div>
```

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
        sess.SetFlash("notice", "🎉 欢迎回来，登录成功！") // CodeIgniter 风格闪存消息，读取后自动销毁
        c.Redirect(302, "/dashboard")
    })

    app.Get("/dashboard", func(c *godeniter.Context, sess session.Session) {
        username := sess.GetString("username")
        notice := sess.GetFlashString("notice") // 仅在本次请求有效
        c.String(200, "欢迎回来: %s, 提示: %s", username, notice)
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

## 📦 跨平台构建、单文件打包与交付部署 (Build & Deploy)

Godeniter 基于 **100% 纯 Go 标准库与 `embed.FS`** 设计，天然具备卓越的跨平台交叉编译能力。在日常开发时，开发者只需专注于业务逻辑编写与 `go run main.go` 调试；在最终向客户机交付部署时，可直接编译为单个二进制文件：

* **运行一键构建脚本（自动化产出全平台二进制）**：
  ```bash
  ./build.sh      # macOS / Linux (一键编译全部应用至 dist/ 目录)
  build.bat       # Windows (一键编译全部应用)
  ```
* **手动交叉编译 Windows 独立 `.exe` 单文件**：
  ```bash
  # 纯 Go 0 依赖，无需配置任何 Windows gcc/CGO 交叉编译环境，秒级生成
  CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o dist/app.exe main.go
  ```
* **纯标准库 Windows 图标编译器与动态检测 (`utils/rsrc` / `cmd/rsrc`)**：
  * **0 外部第三方依赖**：框架原生内置纯 Go 标准库实现的 Windows COFF 资源段编译器，**彻底告别 `rsrc`、`goversioninfo` 或 GCC `windres` 等外部工具链**；
  * **动态检测 `app.ico`**：打包脚本自动探测当前工程是否存在 `app.ico` / `favicon.ico`，按需自动转译为 `resource_windows_amd64.syso`；
  * **桌面级专属图标**：编译出的 Windows `.exe` 单文件在资源管理器与桌面上自动展示定制应用图标，同时作为网站 Favicon 响应，达成专业客户端交付水准。
* **现场免环境交付体验**：
  * 生成的 `.exe` 已经内嵌了前端静态网页与 HTML 模板，无需在客户机安装任何 Go 环境或依赖库；
  * 直接拷贝给客户，**双击即可直接运行**，终端自动输出访问地址与启动 Banner。

👉 详见专有运维手册：[**《跨平台单文件打包与 Windows 客户机交付手册 (docs/build_and_deploy.md)》**](./docs/build_and_deploy.md)

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
  * [`daemon/`](./daemon/) - 跨平台服务生命周期与守护进程管理器（start/stop/restart/status）
  * [`tray/`](./tray/) - 跨平台系统托盘与状态栏常驻客户端模式（macOS/Windows/Linux 0 依赖）
* **通用工具集 (`utils/`)**：
  * [`utils/str/`](./utils/str/) - 字符串处理、脱敏、哈希与随机生成
  * [`utils/upload/`](./utils/upload/) - 文件上传安全处理类与存储校验器
* **命令行与开箱即用示例**：
  * [`cmd/godeniter/`](./cmd/godeniter/) - 官方 CLI 脚手架工具 (`godeniter new`)
  * [`examples/01_api_spa/`](./examples/01_api_spa/) - 架构模式一：前后端分离 RESTful API + SPA 单页 (带文件上传与分页) 完整 Demo
  * [`examples/02_mvc_template/`](./examples/02_mvc_template/) - 架构模式二：经典服务端渲染 MVC + HTML Template 套页面 (带头像上传与搜索) 完整 Demo
  * [`examples/database_mysql/`](./examples/database_mysql/) - 数据库实战工程：MySQL 连接池、CRUD 与事务操作完整示例
  * [`dist/`](./dist/) - 编译生成的跨平台单文件可执行程序输出目录
* **核心文档与开发手册 (`docs/`)**：
  * [`docs/database.md`](./docs/database.md) - 数据库与 ActiveRecord 开发手册 (含 MySQL 生产连接池与 CRUD 实战)
  * [`docs/config.md`](./docs/config.md) - 0 依赖动态配置 (`config.json`)、数据库连接与客户机端口修改手册
  * [`docs/daemon.md`](./docs/daemon.md) - 服务生命周期与守护进程运维手册 (start/stop/restart/status)
  * [`docs/tray.md`](./docs/tray.md) - 跨平台桌面系统托盘与状态栏常驻模式手册 (macOS/Windows/Linux)
  * [`docs/build_and_deploy.md`](./docs/build_and_deploy.md) - 跨平台单文件打包与 Windows 客户机交付手册
  * [`docs/offline.md`](./docs/offline.md) - 离线环境与受限网络开发/编译指南 (Zip 包即用与单文件交付)
  * [`docs/progress.md`](./docs/progress.md) - 框架开发进度、架构设计原则与版本演进记录

---

## 📄 开源许可证 (License)

Godeniter 核心框架基于 **[GNU General Public License v3.0 (GPL-3.0)](./LICENSE)** 协议开源。


# Godeniter 2.0 开发进度与接续指南 (Roadmap & Status)

本文档记录 Godeniter 2.0 的当前开发状态、已完成特性清单以及后续开发建议。任何开发者或 AI 均可依据此文档无缝继续推进项目。

---

## 📅 进度追踪清单 (Progress Checklist)

### ✅ 已完成模块 (Completed)

- [x] **依赖注入容器 (`inject/`)**
  - [x] 基于 `reflect` 的类型映射 (`Map`) 与接口映射 (`MapTo`)
  - [x] 父子级联容器（全局 App -> 请求级 Context）
  - [x] 动态函数依赖解析与调用 (`Invoke`)
  - [x] 结构体字段注入 (`Apply`)
  - [x] 100% 单元测试覆盖 (`inject/inject_test.go`)
- [x] **Trie 树路由器 (`router/`)**
  - [x] RESTful HTTP 动词隔离 (GET/POST/PUT/DELETE/PATCH/OPTIONS/HEAD/ANY)
  - [x] 动态命名参数支持 (`/users/:id`)
  - [x] 通配符全路径捕获 (`/static/*filepath`)
  - [x] 路由多级分组 (`RouterGroup`) 与分组中间件继承
  - [x] 100% 单元测试覆盖 (`router/router_test.go`)
- [x] **Context 上下文与核心引擎 (`godeniter`)**
  - [x] `ResponseWriter` 包装与 HTTP 状态码/写入检测
  - [x] 洋葱圈模型 (`c.Next()`, `c.Abort()`, `c.AbortWithStatusJSON()`)
  - [x] 请求参数获取 (`Param`, `Query`, `PostForm`) 与 JSON 自动绑定 (`BindJSON`)
  - [x] 响应渲染器 (`JSON`, `String`, `HTML`, `Data`)
  - [x] 智能返回值路由解析（支持 Martini 风格返回 `200, string` 或 `200, H{}`）
  - [x] 内置常用中间件：`Logger()`, `Recovery()`, `CORS()`
  - [x] 404 自定义/默认处理
  - [x] 启动 ASCII Banner 与局域网/本机可点击访问链接提示
  - [x] 100% 核心测试覆盖 (`godeniter_test.go`)
- [x] **轻量级 Database 模块 (`db/`)**
  - [x] 标准库 `*sql.DB` 与连接池封装 (`db.Open`)
  - [x] 链式 SQL 构造器 (`Table`, `Select`, `Where`, `OrWhere`, `WhereIn`, `OrderBy`, `GroupBy`, `Limit`, `Offset`)
  - [x] 结构体与切片反射扫描器 (`Find(&slice)`, `First(&struct)`)，支持 `db:"tag"` 与蛇形命名匹配
  - [x] 数据库写操作 (`Insert`, `Update`, `Delete`)
  - [x] 事务封装 (`Transaction(func(tx *db.Tx) error)`)，支持异常自动回滚
  - [x] 100% 单元测试覆盖 (`db/db_test.go`)
- [x] **请求参数绑定与轻量验证器 (`binding/`)**
  - [x] Struct Tag 规则引擎 (`required`, `min`, `max`, `email`, `numeric`)
  - [x] `BindJSON`, `BindQuery`, `BindForm`, `Bind` 多源绑定
  - [x] Context 便捷集成 (`c.BindAndValidate(&dto)`)
  - [x] 100% 单元测试覆盖 (`binding/binding_test.go`)
- [x] **服务端会话管理 (`session/`)**
  - [x] 纯标准库实现，类似 PHP `$_SESSION` 体验
  - [x] 基于 HMAC-SHA256 安全签名的防篡改 `CookieStore`
  - [x] `godeniter.Session` 全自动装配中间件
  - [x] Context 与 Injector 深度集成 (`c.Session()`, `func(sess session.Session)`)
  - [x] CodeIgniter 风格一次性闪存消息 (`SetFlash`, `GetFlash`, `GetFlashString`)
  - [x] 100% 单元测试覆盖 (`session/session_test.go`)
- [x] **生产级平滑优雅停机与 Context 默认值增强**
  - [x] `app.Run` 内置监听系统信号 (`SIGINT`, `SIGTERM`)，支持平滑退出 (Graceful Shutdown)
  - [x] `Context` 便捷默认值读取 (`DefaultQuery`, `DefaultPostForm`, `QueryInt`)
- [x] **通用工具集模块 (`utils/`)**
  - [x] `utils/str/`：字符串随机串、UUID、命名转换、UTF-8 截断、脱敏、哈希与 XSS 过滤
  - [x] `utils/upload/`：文件上传大小/扩展名白名单校验、安全重命名与自动归档
  - [x] 100% 单元测试覆盖 (`utils/str/str_test.go`, `utils/upload/upload_test.go`)
- [x] **官方独立脚手架工程 ([`godeniter-starter`](../godeniter-starter))**
  - [x] 在 `godeniter` 同级创建独立的脚手架项目
  - [x] 规范的 MVC + RESTful API 分层目录骨架（`config/`, `app/controllers/`, `app/models/`, `app/services/`, `app/middleware/`, `views/`）
  - [x] 配置支持从系统环境变量动态覆盖 (`PORT`, `APP_ENV`, `UPLOAD_DIR`, `SESSION_KEY`)
  - [x] 模块依赖配置 (`go.mod` 自动指向 `replace github.com/xbt/godeniter => ../godeniter`)
  - [x] 跨平台单文件打包流水线 (`build.sh`, `build.bat` 输出 `dist/app.exe`)
  - [x] 完整的上手文档 [`README.md`](../godeniter-starter/README.md)
- [x] **官方 CLI 脚手架工具 (`cmd/godeniter/`)**
  - [x] `godeniter new <name> --template=api|mvc` 一键生成全新工程
  - [x] `godeniter version` 与帮助信息输出
- [x] **单文件打包与 Windows 交付支持**
  - [x] `embed.FS` 内嵌 HTML 模板与静态资源
  - [x] 架构模式一：前后端分离 RESTful API + SPA 完整 Demo (`examples/01_api_spa/`)
  - [x] 架构模式二：经典 PHP 风格服务端渲染 MVC 套页面完整 Demo (`examples/02_mvc_template/`)
  - [x] 数据库实战工程：MySQL 连接池与 CRUD 实战示例 (`examples/database_mysql/`)
  - [x] 跨平台一键编译脚本 (`build.sh`, `build.bat`)，一键生成 Windows `.exe`
- [x] **全套 CodeIgniter 风格开发者文档体系**
  - [x] 主入口 [`README.md`](../README.md)（代码用法与单文件打包彻底分离）
  - [x] 架构原理 [`docs/architecture.md`](./architecture.md)
  - [x] 开发进度与接续指南 [`docs/progress.md`](./progress.md)
  - [x] 路由与中间件开发手册 [`docs/routing.md`](./routing.md)
  - [x] 数据库与 ActiveRecord 手册 [`docs/database.md`](./database.md)（含 MySQL 生产连接池与 Web API 实战）
  - [x] 依赖注入与容器手册 [`docs/dependency_injection.md`](./dependency_injection.md)
  - [x] 文件上传与安全存储手册 [`docs/upload.md`](./upload.md)
  - [x] 字符串与安全辅助库手册 [`docs/string_utils.md`](./string_utils.md)
  - [x] 参数绑定与验证器手册 [`docs/binding_validation.md`](./binding_validation.md)
  - [x] 服务端会话管理手册 [`docs/session.md`](./session.md)（已实现 Flash Data 闪存消息）
  - [x] 跨平台单文件打包手册 [`docs/build_and_deploy.md`](./build_and_deploy.md)
  - [x] 离线与受限网络开发指南 [`docs/offline.md`](./offline.md)
  - [x] 动态配置与数据库连接手册 [`docs/config.md`](./config.md)
  - [x] 服务生命周期与守护进程手册 [`docs/daemon.md`](./daemon.md)
  - [x] 跨平台桌面系统托盘开发手册 [`docs/tray.md`](./tray.md)
  - [x] 交互式 CodeIgniter 风格在线 HTML 控制台与用户手册 (`examples/01_api_spa/static/index.html`)
- [x] **无侵入 HTML 模板预处理引擎 (`PreprocessHTMLTemplate`)**
  - [x] 原生支持 `<!--{{ ... }}-->` 注释语法，静态 HTML 原型不乱码，编译后无缝渲染
  - [x] `Engine.SetFuncMap(template.FuncMap)` 链式注册视图辅助函数
  - [x] `LoadHTMLFS` 与 `LoadHTMLGlob` 内置无侵入转译支持
- [x] **服务生命周期与守护进程管理器 (`daemon/`)**
  - [x] 跨平台前台/后台守护模式切换（Setsid / DETACHED_PROCESS）
  - [x] 提供 `start`, `stop`, `restart`, `status` 标准运维指令闭环
  - [x] 自动探测并输出本地与局域网访问地址
- [x] **纯标准库 Windows 图标编译器 (`utils/rsrc`)**
  - [x] 纯 Go 标准库将 `.ico` 转为 COFF `syso`，0 外部工具链依赖
  - [x] `app.ico` 动态探测与单文件打包自动缝合
- [x] **跨平台系统托盘与状态栏客户端模式 (`tray/`)**
  - [x] macOS 状态栏图文并茂自适应宽度与 Accessory 常驻模式
  - [x] Windows 100% 纯 Go 标准库 Win32 托盘驱动，免 CGO，免第三方包
  - [x] 经典原生菜单四件套（打开后台、打开应用目录、关于系统、退出程序）
  - [x] Win32 原生控制台黑框自动隐藏 (`HideConsole`)，托盘模式下彻底免黑框干扰
  - [x] 统一全能单二进制 (`dist/app.exe`)：同一文件兼顾 CLI 运维指令与桌面双击静默托盘
  - [x] 系统信号与主事件循环优雅退出联动
- [x] **纯标准库代码签名证书生成工具 (`cmd/cert`) 与数字签名实战指南**
  - [x] 0 依赖纯 Go 标准库 RSA 2048 + X.509 代码签名专用扩展生成器
  - [x] 自动生成私钥 (`.key`)、DER 格式公钥 (`.cer`)、PEM 证书与标准 PFX 签名包
  - [x] 完整数字签名、UPX 压缩与 Windows 客户机一键信任指南 [`docs/code_signing.md`](./code_signing.md)
  - [x] `.gitignore` 强化机密隔离，严防私钥与证书泄露
- [x] **Trie 树前缀路由深度加固**
  - [x] 静态精准节点与动态参数节点全等隔离，杜绝同级覆盖
  - [x] 严格遵循“静态精准 > 命名参数 > 通配符”搜索优先级


---

## 🔮 待演进规划 (Next Steps & Backlog)

后续接手开发建议优先考虑以下功能：

1. **数据库迁移工具 (Migrator)**：
   * 增加类似 `php artisan migrate` 或 `codeigniter spark migrate` 的 SQL 结构迁移执行器。
2. **连接池健康监测探针 (Health Check)**：
   * 提供针对 MySQL / SQLite 连接状态的实时监测与探针接口。

---

## 🛠️ 常用开发与测试指令

* **运行所有单元测试**：
  ```bash
  go test ./... -v
  ```
* **本地运行 Demo 1 (前后端分离 API + SPA 模式)**：
  ```bash
  go run ./examples/01_api_spa/main.go
  ```
* **本地运行 Demo 2 (经典 MVC 模板渲染模式)**：
  ```bash
  go run ./examples/02_mvc_template/main.go
  ```
* **一键打包全平台单文件可执行程序 (生成 dist/*.exe)**：
  ```bash
  ./build.sh
  ```

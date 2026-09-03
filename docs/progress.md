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
  - [x] 100% 单元测试覆盖 (`session/session_test.go`)
- [x] **通用工具集模块 (`utils/`)**
  - [x] `utils/str/`：字符串随机串、UUID、命名转换、UTF-8 截断、脱敏、哈希与 XSS 过滤
  - [x] `utils/upload/`：文件上传大小/扩展名白名单校验、安全重命名与自动归档
  - [x] 100% 单元测试覆盖 (`utils/str/str_test.go`, `utils/upload/upload_test.go`)
- [x] **官方 CLI 脚手架工具 (`cmd/godeniter/`)**
  - [x] `godeniter new <name> --template=api|mvc` 一键生成全新工程
  - [x] `godeniter version` 与帮助信息输出
- [x] **单文件打包与 Windows 交付支持**
  - [x] `embed.FS` 内嵌 HTML 模板与静态资源
  - [x] 模式一：前后端分离 RESTful API + SPA 完整 Demo (`examples/01_api_spa/`)
  - [x] 模式二：经典 PHP 风格服务端模板渲染 MVC 完整 Demo (`examples/02_mvc_template/`)
  - [x] 跨平台一键编译脚本 (`build.sh`, `build.bat`)，一键生成 Windows `.exe`
- [x] **全套 CodeIgniter 风格开发者文档体系**
  - [x] 主入口 [`README.md`](file:///Users/ben/godeniter/README.md)
  - [x] 架构原理 [`docs/architecture.md`](file:///Users/ben/godeniter/docs/architecture.md)
  - [x] 开发进度与接续指南 [`docs/progress.md`](file:///Users/ben/godeniter/docs/progress.md)
  - [x] 路由与中间件开发手册 [`docs/routing.md`](file:///Users/ben/godeniter/docs/routing.md)
  - [x] 数据库与 ActiveRecord 手册 [`docs/database.md`](file:///Users/ben/godeniter/docs/database.md)
  - [x] 依赖注入与容器手册 [`docs/dependency_injection.md`](file:///Users/ben/godeniter/docs/dependency_injection.md)
  - [x] 文件上传与安全存储手册 [`docs/upload.md`](file:///Users/ben/godeniter/docs/upload.md)
  - [x] 字符串与安全辅助库手册 [`docs/string_utils.md`](file:///Users/ben/godeniter/docs/string_utils.md)
  - [x] 参数绑定与验证器手册 [`docs/binding_validation.md`](file:///Users/ben/godeniter/docs/binding_validation.md)
  - [x] 服务端会话管理手册 [`docs/session.md`](file:///Users/ben/godeniter/docs/session.md)
  - [x] 跨平台单文件打包手册 [`docs/build_and_deploy.md`](file:///Users/ben/godeniter/docs/build_and_deploy.md)
  - [x] 交互式 CodeIgniter 风格在线 HTML 控制台与用户手册 (`examples/01_api_spa/static/index.html`)

---

## 🔮 待演进规划 (Next Steps & Backlog)

后续接手开发建议优先考虑以下功能：

1. **数据库驱动接入样例**：
   * 在 `examples/` 目录下创建连接 SQLite (`modernc.org/sqlite`) 或 MySQL 的完整 CRUD 业务模块。
2. **Flash 消息支持**：
   * 在 session 基础上增加类似 CodeIgniter 的一次性提示消息（Flash Data）。
3. **数据库迁移工具 (Migrator)**：
   * 增加类似 `php artisan migrate` 的 SQL 迁移执行器。

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

# Godeniter 2.0 架构设计与实现文档

本文档用于详细阐述 Godeniter 2.0 框架的设计哲学、内部机制与各模块协作原理。供后续接手的开发者或 AI 快速掌握全貌。

---

## 一、 设计哲学与目标

1. **0 外部依赖 (Zero Dependencies Core)**：
   * 整个 Web 框架核心（DI 注入器、Trie 路由器、Context 上下文、洋葱圈中间件、QueryBuilder 链式构建器）**100% 基于 Go 标准库实现**，杜绝庞杂的三方库供应链隐患与版本冲突。
2. **Martini 风格的极简开发体验**：
   * 借鉴 Martini（codegangsta/inject）的依赖注入思想，Handler 签名不拘一格，可自由声明参数。
   * 同时保留原生 `func(c *Context)` 签名，兼顾灵活性与高性能。
3. **PHP (CodeIgniter / Laravel) 风格的数据库操作**：
   * 在标准库 `database/sql` 抽象之上封装类似 ActiveRecord 的链式调用，使得 PHP 背景开发者零学习成本上手。
4. **单文件交付与跨平台部署**：
   * 利用 Go 原生 `embed` 特性内嵌 HTML 视图与静态资源，一键编译为 `.exe`，双击直接运行。

---

## 二、 核心模块架构

```
┌────────────────────────────────────────────────────────┐
│                   Engine (godeniter.go)                │
│    - 组合 RouterGroup (路由API)                          │
│    - 组合 Injector (全局依赖容器)                         │
│    - 实现 http.Handler (ServeHTTP)                     │
└──────────┬─────────────────────────────┬───────────────┘
           │                             │
           ▼                             ▼
┌──────────────────────┐      ┌──────────────────────────┐
│   Trie 树路由器      │      │     依赖注入容器 (DI)    │
│  (router/trie.go)    │      │    (inject/inject.go)    │
│  - 动态参数 :param   │      │  - Type / Interface 映射 │
│  - 通配符 *filepath  │      │  - Invoke 动态参数解析   │
│  - RESTful 动词隔离  │      │  - 父子容器继承          │
└──────────┬───────────┘      └──────────┬───────────────┘
           │                             │
           └──────────────┬──────────────┘
                          │
                          ▼
             ┌─────────────────────────┐
             │   Context (context.go)  │
             │  - 洋葱圈 Next / Abort  │
             │  - 参数获取与 JSON 绑定 │
             │  - HTML / JSON 智能渲染 │
             └─────────────────────────┘
```

---

## 三、 核心模块实现细节

### 1. 依赖注入模块 (`inject/`)
* **`Map(val interface{})`**：按 concrete type 映射到容器字典中。
* **`MapTo(val interface{}, ifacePtr interface{})`**：按接口类型映射（例如将具体的实现注入到接口类型）。
* **`Invoke(f interface{})`**：通过 `reflect.TypeOf(f)` 提取参数类型列表，逐一在容器（及递归在父容器）中查找匹配的值，通过 `reflect.ValueOf(f).Call(args)` 调用。
* **`Apply(val interface{})`**：反射解析结构体带有 `inject:""` 标签的字段并自动赋值。

### 2. Trie 树路由器 (`router/`)
* **节点定义 (`node`)**：包含 `part`（当前分段）、`children`（子节点列表）、`isWild`（是否为 `:id` 或 `*filepath`）、`pattern`（路由模式）和 `handlers`（中间件及处理函数列表）。
* **匹配优先级**：优先匹配精准路径片段，未找到时匹配动态通配节点。
* **参数提取**：路由匹配成功时，将 URL 中的实际值与注册模式中的 `:param` 名字对齐，存入 `router.Params`。

### 3. 请求上下文与中间件执行 (`context.go`)
* **洋葱圈模型**：
  * 每个 Context 维护一个 `index` 游标和 `handlers` 切片。
  * 调用 `c.Next()` 时 `index++` 递归推进调用下一个 Handler。
  * 调用 `c.Abort()` 时将 `index` 跳至最大值，阻止后续执行。
* **智能 Handler 分发**：
  * 若为 `func(*Context)` 或 `HandlerFunc`，直接执行（零反射，极速）。
  * 若为普通函数，自动调用 `c.Invoke(fn)` 进行参数解析并调用，同时根据函数返回值智能输出（例如返回 `int, string` 自动调用 `c.String(code, str)`，返回 `int, H` 自动作为 JSON 响应）。

### 4. 数据库链式构造器 (`db/`)
* 基于 Go 标准库 `database/sql`。
* 链式拼接 SQL 语句与参数切片（`ToSQL()`），避免 SQL 注入风险。
* 通过 `db/scanner.go` 中的反射映射，自动将 `*sql.Rows` 扫描到结构体切片或单条结构体（支持 `db:"tag"` 映射或自动驼峰下划线转换）。

### 5. 请求参数绑定与轻量验证器 (`binding/`)
* 基于 Go 标准库反射解析 Struct Tag（`binding:"required,min=4,max=16,email,numeric"`）。
* 支持 `BindJSON`、`BindQuery`、`BindForm` 及智能匹配的 `Bind` 方法。
* 提供清晰精确的 `ValidationError` 与 `FieldError`，可直接通过 `c.BindAndValidate(&dto)` 实现一行代码解析与校验。

### 6. 服务端会话管理 (`session/`)
* 纯标准库实现，提供类似 PHP `$_SESSION` 的极简体验。
* 内置 `CookieStore`，基于 HMAC-SHA256 签名机制防篡改，非常适合无状态单文件分布式/免配置交付。
* 通过 `godeniter.Session(...)` 中间件自动装配并注入到 Context 与 Injector 中。

### 7. 通用工具库体系 (`utils/`)
* **`utils/str/`**：类似 PHP 经典辅助函数，提供高强度随机串生成、UUIDv4、大小驼峰/蛇形/短横线互转、UTF-8 字符截断、敏感信息脱敏（手机号/邮箱/身份证）与安全哈希。
* **`utils/upload/`**：对标 CodeIgniter Upload 类，提供上传文件安全存储、大小限制与扩展名白名单拦截器。

### 8. 纯标准库 Windows 图标编译器 (`utils/rsrc` / `cmd/rsrc`)
* **0 外部工具依赖**：框架内原生实现了将 `.ico` 文件解析并转译为 Windows COFF `resource_windows_amd64.syso` 的纯 Go 编译器，彻底替代第三方的 `rsrc` 或 GCC `windres`。
* **动态缝合**：打包脚本自动探测项目中的 `app.ico`，在构建时将资源段缝合进 `.exe`，实现双端（Windows 桌面图标 + Web Favicon）一体化。

### 9. 服务生命周期与守护进程管理 (`daemon/`)
* **纯 Go 跨平台实现**：在 Linux/macOS 上通过系统调用脱离终端（Setsid），在 Windows 上通过 `DETACHED_PROCESS` 剥离黑框控制台。
* **类似 Nginx 指令**：统一提供 `start`, `stop`, `restart`, `status` 与 `app.pid` / `app.log` 状态与重定向闭环管理。

### 10. 跨平台桌面系统托盘与状态栏 (`tray/`)
* **macOS 端**：基于原生 Cocoa / AppKit 驱动，以 Accessory 模式常驻右上角状态栏，自适应宽度图文并茂，支持经典四件套下拉菜单。
* **Windows 端**：100% 纯 Go 标准库 `syscall` Win32 API（`shell32.dll` 与 `user32.dll`）驱动，0 CGO 免编译门槛；遵循 KB139526 规范避免二次右键失灵。
* **优雅停机协同**：捕获系统退出信号，实现点击退出或终端 Ctrl+C 时安全平滑关闭 HTTP 服务并秒级退出。

---

## 四、 后续演进路线建议

1. **数据库迁移工具 (Migrator)**：增加类似 `php artisan migrate` 的 SQL 结构迁移执行器。
2. **连接池实时健康探针 (Health Check)**：提供面向生产微服务探针的 HTTP 健康检查端点。

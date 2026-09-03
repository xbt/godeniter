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

---

## 四、 后续演进路线建议

1. **数据库驱动示例**：可按需编写接入纯 Go SQLite (`modernc.org/sqlite`) 或 MySQL (`github.com/go-sql-driver/mysql`) 的完整增删改查业务模块。
2. **Session / Cookie 管理**：增加标准库实现的轻量签名 Cookie / 内存 Session 支持。
3. **验证器 (Validator)**：增加基于 Struct Tag 的数据校验工具（如 `binding:"required,min=6"`）。

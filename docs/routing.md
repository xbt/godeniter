# 路由与中间件开发手册 (Routing & Middleware)

Godeniter 内置了基于前缀树（Trie Tree）的高性能动态路由器，支持 RESTful 动词、参数捕获、全路径通配符、多级分组以及洋葱圈中间件。

---

## 1. 基础路由注册

```go
package main

import (
    "github.com/xbt/godeniter"
    "net/http"
)

func main() {
    app := godeniter.New()

    // 1. 标准 Context 签名
    app.Get("/hello", func(c *godeniter.Context) {
        c.String(http.StatusOK, "Hello Godeniter!")
    })

    // 2. 支持所有标准 HTTP 动词
    app.Post("/articles", createArticleHandler)
    app.Put("/articles/:id", updateArticleHandler)
    app.Delete("/articles/:id", deleteArticleHandler)
    app.Patch("/articles/:id", patchArticleHandler)
    app.Any("/webhook", anyWebhookHandler) // 匹配任意 HTTP 动词

    app.Run(":8080")
}
```

---

## 2. 动态路由与通配符参数

| 模式语法 | 示例 URL | 匹配效果 | 获取方式 |
| :--- | :--- | :--- | :--- |
| `/users/:id` | `/users/1001` | 匹配单级动态参数 | `c.Param("id")` -> `"1001"` |
| `/static/*filepath` | `/static/css/app.css` | 捕获后续所有路径 | `c.Param("filepath")` -> `"css/app.css"` |

### 代码示例：
```go
// 方式一：通过 Context 获取
app.Get("/users/:id", func(c *godeniter.Context) {
    userID := c.Param("id")
    c.JSON(200, godeniter.H{"user_id": userID})
})

// 方式二：Martini 风格直接注入 router.Params
app.Get("/posts/:category/:slug", func(params router.Params) (int, godeniter.H) {
    return 200, godeniter.H{
        "category": params.Get("category"),
        "slug":     params.Get("slug"),
    }
})
```

---

## 3. 路由分组与多级嵌套 (RouterGroup)

类似 CodeIgniter 4 或 Laravel 的路由组功能：

```go
api := app.Group("/api")
{
    v1 := api.Group("/v1")
    {
        v1.Get("/users", listUsers)
        v1.Get("/users/:id", getUserDetail)
    }

    v2 := api.Group("/v2")
    {
        v2.Get("/users", listUsersV2)
    }
}
```

---

## 4. 中间件与洋葱圈模型 (Middleware)

中间件签名均为 `func(*godeniter.Context)` 或任意支持依赖注入的函数。通过 `c.Next()` 执行后续链条，通过 `c.Abort()` 终止请求。

### 自定义鉴权中间件示例：
```go
func AuthRequired() godeniter.HandlerFunc {
    return func(c *godeniter.Context) {
        token := c.Req.Header.Get("Authorization")
        if token == "" {
            c.Fail(40101, "未提供认证 Token，无权访问")
            c.Abort() // 阻断后续 Handler 执行
            return
        }
        
        // 传递用户信息到上下文
        c.Set("current_user", "Admin")
        c.Next() // 放行执行后续业务
    }
}

// 全局应用
app.Use(godeniter.Logger(), godeniter.Recovery())

// 仅在指定路由组应用
admin := app.Group("/admin", AuthRequired())
{
    admin.Get("/dashboard", dashboardHandler)
}
```

---

## 5. 自定义 404 页面与未命中处理

当请求的路径未匹配到任何路由规则时，默认返回纯文本 404。可通过覆盖 `app.NotFound` 自定义返回 JSON 格式或渲染友好 HTML 页面：

```go
// 1. API 模式返回统一业务 JSON
app.NotFound = func(c *godeniter.Context) {
    c.Fail(40400, fmt.Sprintf("路由 [%s %s] 不存在", c.Method, c.Path))
}

// 2. 或渲染友好的 404 HTML 模板
app.NotFound = func(c *godeniter.Context) {
    c.HTML(404, "404.html", godeniter.H{
        "Title": "页面走丢了",
    })
}
```

---

## 6. 参数获取与默认值提取

除了通过 `c.Param("id")` 获取路径变量外，还提供了带 fallback 默认值的安全提取方法：

```go
// 1. 获取 URL Query 参数 (若未传则返回默认值)
keyword := c.DefaultQuery("keyword", "默认搜索词")
page := c.QueryInt("page", 1)           // 自动转为 int，若无效或未传返回 1

// 2. 获取 POST 表单数据 (若未传则返回默认值)
role := c.DefaultPostForm("role", "guest")
```

---

## 7. 生产级平滑优雅退出 (Graceful Shutdown)

`app.Run(":8080")` 内部已内置监听操作系统终止信号（`os.Interrupt`、`syscall.SIGTERM`）。
当在控制台按下 `Ctrl+C` 或运维执行 `kill` 时，服务不会瞬间截断网络连接，而是给予最多 5 秒超时，等待正在执行的文件上传、数据库事务处理完毕后再安全退出。

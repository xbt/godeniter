# 路由与中间件开发手册 (Routing & Middleware)

Godeniter 内置了基于前缀树（Trie Tree）的高性能动态路由器，支持 RESTful 动词、参数捕获、全路径通配符、多级分组以及洋葱圈中间件。

---

## 1. 基础路由注册

```go
package main

import (
    "godeniter"
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

# 服务端会话管理手册 (Session Management)

对标 **PHP 原生 `$_SESSION`** 与 **CodeIgniter Session Library**。

---

## 1. 快速上手

```go
package main

import (
    "github.com/xbt/godeniter"
    "github.com/xbt/godeniter/session"
)

func main() {
    app := godeniter.Classic()

    // 1. 初始化 HMAC-SHA256 安全签名 Cookie 存储器
    store := session.NewCookieStore("my-secret-key-123456")
    app.Use(godeniter.Session(store, "app_session"))

    // 2. 写入 Session (登录)
    app.Post("/login", func(c *godeniter.Context, sess session.Session) {
        sess.Set("user_id", 1001)
        sess.Set("username", "admin")
        c.Redirect(302, "/dashboard")
    })

    // 3. 读取 Session
    app.Get("/dashboard", func(c *godeniter.Context, sess session.Session) {
        username := sess.GetString("username")
        c.String(200, "当前登录用户: %s", username)
    })

    // 4. 清除 Session (注销)
    app.Get("/logout", func(c *godeniter.Context, sess session.Session) {
        sess.Delete("user_id")
        sess.Delete("username")
        c.Redirect(302, "/")
    })

    app.Run(":8080")
}
```

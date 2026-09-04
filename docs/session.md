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

---

## 2. 一次性闪存消息 (FlashData)

对标 **CodeIgniter `$this->session->set_flashdata()`**：
在用户执行表单提交、修改或注销操作后，通常需要重定向并在目标页面展示一次性通知（如“发布成功”、“密码修改完成”）。FlashData 在被读取后会自动销毁，刷新页面不再显示。

```go
// 1. 设置一次性消息并重定向
app.Post("/articles", func(c *godeniter.Context, sess session.Session) {
    // 业务创建文章...
    sess.SetFlash("notice", "🎉 文章发布成功！")
    c.Redirect(302, "/articles")
})

// 2. 目标页面读取闪存消息 (读取后自动销毁)
app.Get("/articles", func(c *godeniter.Context, sess session.Session) {
    flashMsg := sess.GetFlashString("notice") // 获取后自动从 Session 中删除
    c.HTML(200, "articles.html", godeniter.H{
        "Notice": flashMsg,
    })
})
```

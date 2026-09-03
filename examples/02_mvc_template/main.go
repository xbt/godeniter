package main

import (
	"embed"
	"godeniter"
	"godeniter/examples/02_mvc_template/controllers"
	"godeniter/session"
	"html/template"
	"io/fs"
	"os"
)

//go:embed views/*
var viewsFS embed.FS

func main() {
	// 确保存储目录存在
	_ = os.MkdirAll("./uploads/avatars", 0755)

	// 1. 初始化经典引擎
	app := godeniter.Classic()

	// 2. 静态文件路由
	app.Static("/uploads", "./uploads")

	// 3. 挂载会话管理中间件 (基于 HMAC 签名防篡改 CookieStore)
	store := session.NewCookieStore("godeniter-mvc-secret-key-123456")
	app.Use(godeniter.Session(store, "mvc_session"))

	// 4. 加载嵌入式 HTML 模板 (单文件打包)
	subViews, _ := fs.Sub(viewsFS, "views")
	app.SetHTMLTemplate(template.Must(template.ParseFS(subViews, "*.html")))

	// 5. 实例化控制器
	homeCtrl := &controllers.HomeController{}
	userCtrl := &controllers.UserController{}

	// 6. 路由注册 (类似 CodeIgniter 路由映射)
	app.Get("/", homeCtrl.Index)
	app.Post("/upload/avatar", homeCtrl.UploadAvatar)
	app.Get("/login", userCtrl.LoginForm)
	app.Post("/login", userCtrl.LoginSubmit)
	app.Get("/logout", userCtrl.Logout)

	// 7. 启动 HTTP 服务 (默认监听 8080)
	_ = app.Run(":8080")
}

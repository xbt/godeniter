package main

import (
	"embed"
	"godeniter"
	"godeniter/examples/02_mvc_template/controllers"
	"html/template"
	"io/fs"
)

//go:embed views/*
var viewsFS embed.FS

func main() {
	// 1. 初始化经典引擎
	app := godeniter.Classic()

	// 2. 加载嵌入式 HTML 模板 (单文件打包，无需携带外部 views 文件夹)
	subViews, _ := fs.Sub(viewsFS, "views")
	app.SetHTMLTemplate(template.Must(template.ParseFS(subViews, "*.html")))

	// 3. 实例化控制器
	homeCtrl := &controllers.HomeController{}
	userCtrl := &controllers.UserController{}

	// 4. 路由注册 (类似 CodeIgniter 路由映射)
	app.Get("/", homeCtrl.Index)
	app.Get("/login", userCtrl.LoginForm)
	app.Post("/login", userCtrl.LoginSubmit)
	app.Get("/logout", userCtrl.Logout)

	// 5. 启动 HTTP 服务 (默认监听 8080)
	_ = app.Run(":8080")
}

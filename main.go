package main

import (
	"godeniter/controllers"
	"net/http"
)

func main() {
	// 加载和初始化插件
	//loadPlugins()
	initializePlugins()

	// 创建控制器实例
	homeController := &controllers.HomeController{}

	// 注册控制器中的路由
	registerRoutes(homeController)

	// 启动服务器
	http.ListenAndServe(":8080", router())

	// 执行插件
	//executePlugins()
}

func registerRoutes(homeController *controllers.HomeController) {
	addRoute("/", homeController.Index)
	addRoute("/hello", homeController.Hello)
}

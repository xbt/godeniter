package main

import (
	"embed"
	"godeniter"
	"godeniter/examples/01_api_spa/handlers"
	"godeniter/middleware"
	"net/http"
)

//go:embed static/*
var staticFS embed.FS

func main() {
	// 1. 初始化带有 Logger 和 Recovery 的经典引擎
	app := godeniter.Classic()

	// 2. 启用跨域中间件 (CORS)
	app.Use(middleware.CORS())

	// 3. 依赖注入：注入文章业务服务实例 (单例模式)
	articleService := handlers.NewArticleService()
	app.Map(articleService)

	// 4. 内嵌静态页面路由 (根路径直接提供 SPA 页面)
	app.Get("/", func(c *godeniter.Context) {
		indexData, err := staticFS.ReadFile("static/index.html")
		if err != nil {
			c.String(http.StatusInternalServerError, "加载静态资源失败")
			return
		}
		c.Data(http.StatusOK, "text/html; charset=utf-8", indexData)
	})

	// 5. RESTful API 路由分组 (/api/v1)
	api := app.Group("/api/v1")
	{
		// GET /api/v1/articles        -> 获取文章列表
		api.Get("/articles", handlers.ListArticles)

		// GET /api/v1/articles/:id    -> 获取单篇文章详情 (Martini 风格 DI 演示)
		api.Get("/articles/:id", handlers.GetArticleDetail)

		// POST /api/v1/articles       -> 新增文章 (JSON 绑定)
		api.Post("/articles", handlers.CreateArticle)

		// DELETE /api/v1/articles/:id -> 删除文章
		api.Delete("/articles/:id", handlers.DeleteArticle)
	}

	// 6. 启动服务器 (默认监听 8080 端口)
	// 客户双击 .exe 后即可在终端看到访问地址
	_ = app.Run(":8080")
}

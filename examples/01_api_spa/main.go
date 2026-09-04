package main

import (
	"embed"
	"github.com/xbt/godeniter"
	"github.com/xbt/godeniter/examples/01_api_spa/handlers"
	"github.com/xbt/godeniter/middleware"
	"net/http"
	"os"
)

//go:embed static/*
var staticFS embed.FS

func main() {
	// 确保本地上传存储目录存在
	_ = os.MkdirAll("./uploads/images", 0755)

	// 1. 初始化经典引擎
	app := godeniter.Classic()

	// 2. 启用跨域中间件 (CORS)
	app.Use(middleware.CORS())

	// 3. 依赖注入：注入文章业务服务实例
	articleService := handlers.NewArticleService()
	app.Map(articleService)

	// 4. 静态资源映射：使上传的文件可在浏览器直接访问
	app.Static("/uploads", "./uploads")

	// 5. 内嵌静态单页 SPA 路由
	app.Get("/", func(c *godeniter.Context) {
		indexData, err := staticFS.ReadFile("static/index.html")
		if err != nil {
			c.String(http.StatusInternalServerError, "加载静态资源失败")
			return
		}
		c.Data(http.StatusOK, "text/html; charset=utf-8", indexData)
	})

	// 6. RESTful API 路由分组 (/api/v1)
	api := app.Group("/api/v1")
	{
		// 文章 CRUD 与分页检索
		api.Get("/articles", handlers.ListArticles)
		api.Get("/articles/:id", handlers.GetArticleDetail)
		api.Post("/articles", handlers.CreateArticle)
		api.Delete("/articles/:id", handlers.DeleteArticle)

		// 文件上传接口 (支持图片上传与安全校验)
		api.Post("/upload", handlers.UploadFile)
	}

	// 7. 启动服务器
	_ = app.Run(":8080")
}

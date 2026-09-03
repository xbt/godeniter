package main

import (
	"embed"
	"godeniter"
	"godeniter/router"
	"html/template"
	"io/fs"
	"net/http"
	"time"
)

//go:embed views/*
var viewsFS embed.FS

// SystemInfo 定义系统配置与状态信息（用于依赖注入演示）
type SystemInfo struct {
	AppName   string
	Version   string
	StartTime string
}

func main() {
	// 1. 创建带有 Logger 和 Recovery 的经典引擎
	app := godeniter.Classic()

	// 2. 注入全局服务/对象 (类似 Martini / Laravel 容器)
	sysInfo := &SystemInfo{
		AppName:   "Godeniter Demo App",
		Version:   "2.0.0",
		StartTime: time.Now().Format("2006-01-02 15:04:05"),
	}
	app.Map(sysInfo)

	// 3. 加载内嵌的 HTML 模板 (单文件打包，无需携带外部 views 文件夹)
	subViews, _ := fs.Sub(viewsFS, "views")
	app.SetHTMLTemplate(template.Must(template.ParseFS(subViews, "*.html")))

	// 4. 路由定义：首页 (渲染内嵌模板)
	app.Get("/", func(c *godeniter.Context, info *SystemInfo) {
		c.HTML(http.StatusOK, "index.html", godeniter.H{
			"AppName":   info.AppName,
			"Version":   info.Version,
			"StartTime": info.StartTime,
			"Time":      time.Now().Format("15:04:05"),
		})
	})

	// 5. Martini 风格路由：参数动态注入 + 返回值自动序列化为 JSON
	app.Get("/api/info", func(info *SystemInfo) (int, godeniter.H) {
		return http.StatusOK, godeniter.H{
			"status":     "success",
			"app_name":   info.AppName,
			"version":    info.Version,
			"start_time": info.StartTime,
			"timestamp":  time.Now().Unix(),
		}
	})

	// 6. 动态路由参数演示
	app.Get("/hello/:name", func(params router.Params) string {
		name := params.Get("name")
		return "Hello, " + name + "! Welcome to Godeniter 2.0!\n"
	})

	// 7. 路由分组 API
	api := app.Group("/api/v1")
	{
		type GreetRequest struct {
			Message string `json:"message"`
		}

		api.Post("/echo", func(c *godeniter.Context) {
			var req GreetRequest
			if err := c.BindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, godeniter.H{"error": "无效的 JSON 数据"})
				return
			}
			c.JSON(http.StatusOK, godeniter.H{
				"echo":      req.Message,
				"received":  true,
				"server_at": time.Now().Format("2006-01-02 15:04:05"),
			})
		})
	}

	// 8. 启动服务器并在控制台输出访问地址 (默认 8080 端口)
	// 客户在 Windows 上双击运行 .exe 时，即可在控制台看到可访问的 URL 地址
	_ = app.Run(":8080")
}

// Package main 提供了 Godeniter 框架的官方命令行脚手架工具。
// 类似于 php artisan 或 codeigniter spark，支持一键创建前后端分离 API 与经典 MVC 新工程模版。
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const version = "2.0.0"

func main() {
	if len(os.Args) < 2 {
		printHelp()
		return
	}

	command := os.Args[1]
	switch command {
	case "new":
		handleNewCommand(os.Args[2:])
	case "version", "-v", "--version":
		fmt.Printf("Godeniter CLI Framework Scaffolder v%s\n", version)
	case "help", "-h", "--help":
		printHelp()
	default:
		fmt.Printf("未知指令: %s\n\n", command)
		printHelp()
	}
}

func printHelp() {
	banner := `
   ______          __            _ __            
  / ____/___  ____/ /__  ____   (_) /____  _____ 
 / / __/ __ \/ __  / _ \/ __ \ / / __/ _ \/ ___/ 
/ /_/ / /_/ / /_/ /  __/ / / // / /_/  __/ /     
\____/\____/\__,_/\___/_/ /_//_/\__/\___/_/      
`
	fmt.Print(banner)
	fmt.Printf("Godeniter 官方项目脚手架工具 (v%s)\n\n", version)
	fmt.Println("用法:")
	fmt.Println("  godeniter <command> [arguments]")
	fmt.Println("\n可用指令:")
	fmt.Println("  new <project-name> [--template=api|mvc]   一键初始化全新 Godeniter 项目")
	fmt.Println("  version                                   查看 CLI 与框架版本")
	fmt.Println("  help                                      查看帮助信息")
	fmt.Println("\n模版选项 (--template):")
	fmt.Println("  api (默认)  前后端分离 RESTful API + SPA 单页内嵌方案")
	fmt.Println("  mvc         经典 PHP 风格服务端模板渲染 (MVC + HTML Template) 方案")
	fmt.Println("\n示例:")
	fmt.Println("  godeniter new my-api-app --template=api")
	fmt.Println("  godeniter new my-web-app --template=mvc")
}

func handleNewCommand(args []string) {
	if len(args) < 1 {
		fmt.Println("错误: 请指定新项目名称。示例: godeniter new my-app")
		return
	}

	projectName := args[0]
	fs := flag.NewFlagSet("new", flag.ExitOnError)
	templateType := fs.String("template", "api", "项目模版类型 (api 或 mvc)")
	_ = fs.Parse(args[1:])

	targetDir := filepath.Clean(projectName)
	if _, err := os.Stat(targetDir); err == nil {
		fmt.Printf("错误: 目标目录 [%s] 已存在，请更换项目名或清理目录后重试。\n", targetDir)
		return
	}

	fmt.Printf("🚀 正在创建新项目: %s (模版: %s)...\n", projectName, *templateType)

	var err error
	if *templateType == "mvc" {
		err = generateMVCTemplate(targetDir, projectName)
	} else {
		err = generateAPITemplate(targetDir, projectName)
	}

	if err != nil {
		fmt.Printf("❌ 创建项目失败: %v\n", err)
		return
	}

	fmt.Println("==========================================================")
	fmt.Printf("🎉 项目 [%s] 创建成功！\n", projectName)
	fmt.Println("==========================================================")
	fmt.Println("快速开始:")
	fmt.Printf("  cd %s\n", projectName)
	fmt.Println("  go mod tidy")
	fmt.Println("  go run main.go")
	fmt.Println("\n打包单文件 (Windows 双击运行):")
	fmt.Println("  ./build.sh  或  build.bat")
	fmt.Println("==========================================================")
}

func generateAPITemplate(dir, projectName string) error {
	dirs := []string{
		dir,
		filepath.Join(dir, "handlers"),
		filepath.Join(dir, "models"),
		filepath.Join(dir, "static"),
	}
	for _, d := range dirs {
		if err := os.MkdirAll(d, 0755); err != nil {
			return err
		}
	}

	// 1. go.mod
	goModContent := fmt.Sprintf(`module %s

go 1.20

require github.com/xbt/godeniter v0.0.0
replace github.com/xbt/godeniter => ../godeniter
`, projectName)
	writeFile(filepath.Join(dir, "go.mod"), goModContent)

	// 2. main.go
	mainGoContent := `package main

import (
	"embed"
	"github.com/xbt/godeniter"
	"github.com/xbt/godeniter/middleware"
	"` + projectName + `/handlers"
	"net/http"
)

//go:embed static/*
var staticFS embed.FS

func main() {
	app := godeniter.Classic()
	app.Use(middleware.CORS())

	// 注入业务服务
	articleService := handlers.NewArticleService()
	app.Map(articleService)

	// 根路径提供 SPA 页面
	app.Get("/", func(c *godeniter.Context) {
		indexData, _ := staticFS.ReadFile("static/index.html")
		c.Data(http.StatusOK, "text/html; charset=utf-8", indexData)
	})

	// API 路由
	api := app.Group("/api/v1")
	{
		api.Get("/articles", handlers.ListArticles)
		api.Post("/articles", handlers.CreateArticle)
	}

	_ = app.Run(":8080")
}
`
	writeFile(filepath.Join(dir, "main.go"), mainGoContent)

	// 3. handlers/article.go
	handlerContent := `package handlers

import (
	"godeniter"
	"` + projectName + `/models"
	"sync"
	"time"
)

type ArticleService struct {
	mu       sync.RWMutex
	articles []*models.Article
}

func NewArticleService() *ArticleService {
	return &ArticleService{
		articles: []*models.Article{
			{ID: 1, Title: "欢迎使用 Godeniter 2.0 API 模式", Author: "Admin", CreatedAt: time.Now()},
		},
	}
}

func (s *ArticleService) List() []*models.Article {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.articles
}

func (s *ArticleService) Add(title, author string) *models.Article {
	s.mu.Lock()
	defer s.mu.Unlock()
	a := &models.Article{
		ID:        len(s.articles) + 1,
		Title:     title,
		Author:    author,
		CreatedAt: time.Now(),
	}
	s.articles = append(s.articles, a)
	return a
}

func ListArticles(c *godeniter.Context, svc *ArticleService) {
	c.Success(godeniter.H{
		"items": svc.List(),
	})
}

func CreateArticle(c *godeniter.Context, svc *ArticleService) {
	var req models.CreateArticleRequest
	if err := c.BindAndValidate(&req); err != nil {
		c.Fail(40001, err.Error())
		return
	}
	a := svc.Add(req.Title, req.Author)
	c.Success(a)
}
`
	writeFile(filepath.Join(dir, "handlers", "article.go"), handlerContent)

	// 4. models/article.go
	modelContent := `package models

import "time"

type Article struct {
	ID        int       ` + "`json:\"id\"`" + `
	Title     string    ` + "`json:\"title\"`" + `
	Author    string    ` + "`json:\"author\"`" + `
	CreatedAt time.Time ` + "`json:\"created_at\"`" + `
}

type CreateArticleRequest struct {
	Title  string ` + "`json:\"title\" binding:\"required,min=4\"`" + `
	Author string ` + "`json:\"author\" binding:\"required\"`" + `
}
`
	writeFile(filepath.Join(dir, "models", "article.go"), modelContent)

	// 5. static/index.html
	htmlContent := `<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>` + projectName + ` - API Dashboard</title>
    <style>body { background:#0f172a; color:#f8fafc; font-family:sans-serif; padding:40px; text-align:center; }</style>
</head>
<body>
    <h1>🚀 ` + projectName + ` 正在运行</h1>
    <p>前后端分离 RESTful API 架构已就绪。</p>
    <p><a href="/api/v1/articles" style="color:#38bdf8;">查看 API 接口 &rarr;</a></p>
</body>
</html>`
	writeFile(filepath.Join(dir, "static", "index.html"), htmlContent)

	// 6. build.sh
	buildSh := `#!/bin/bash
mkdir -p dist
CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o dist/app.exe main.go
go build -ldflags="-s -w" -o dist/app main.go
echo "Build completed in dist/"
`
	writeFile(filepath.Join(dir, "build.sh"), buildSh)
	_ = os.Chmod(filepath.Join(dir, "build.sh"), 0755)

	return nil
}

func generateMVCTemplate(dir, projectName string) error {
	dirs := []string{
		dir,
		filepath.Join(dir, "controllers"),
		filepath.Join(dir, "models"),
		filepath.Join(dir, "views"),
	}
	for _, d := range dirs {
		if err := os.MkdirAll(d, 0755); err != nil {
			return err
		}
	}

	// 1. go.mod
	goModContent := fmt.Sprintf(`module %s

go 1.20

require github.com/xbt/godeniter v0.0.0
replace github.com/xbt/godeniter => ../godeniter
`, projectName)
	writeFile(filepath.Join(dir, "go.mod"), goModContent)

	// 2. main.go
	mainGoContent := `package main

import (
	"embed"
	"github.com/xbt/godeniter"
	"html/template"
	"io/fs"
	"` + projectName + `/controllers"
)

//go:embed views/*
var viewsFS embed.FS

func main() {
	app := godeniter.Classic()

	subViews, _ := fs.Sub(viewsFS, "views")
	app.SetHTMLTemplate(template.Must(template.ParseFS(subViews, "*.html")))

	homeCtrl := &controllers.HomeController{}
	app.Get("/", homeCtrl.Index)

	_ = app.Run(":8080")
}
`
	writeFile(filepath.Join(dir, "main.go"), mainGoContent)

	// 3. controllers/home.go
	ctrlContent := `package controllers

import (
	"github.com/xbt/godeniter"
	"net/http"
)

type HomeController struct{}

func (ctrl *HomeController) Index(c *godeniter.Context) {
	c.HTML(http.StatusOK, "index.html", godeniter.H{
		"Title": "` + projectName + ` - MVC 控制台",
	})
}
`
	writeFile(filepath.Join(dir, "controllers", "home.go"), ctrlContent)

	// 4. views/index.html
	htmlContent := `<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>{{ .Title }}</title>
    <style>body { background:#f1f5f9; color:#1e293b; font-family:sans-serif; padding:40px; text-align:center; }</style>
</head>
<body>
    <h1>🚀 {{ .Title }} 已成功启动！</h1>
    <p>基于 PHP 风格服务端渲染 (SSR) 驱动，单文件内嵌打包。</p>
</body>
</html>`
	writeFile(filepath.Join(dir, "views", "index.html"), htmlContent)

	// 5. build.sh
	buildSh := `#!/bin/bash
mkdir -p dist
CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o dist/app.exe main.go
go build -ldflags="-s -w" -o dist/app main.go
echo "Build completed in dist/"
`
	writeFile(filepath.Join(dir, "build.sh"), buildSh)
	_ = os.Chmod(filepath.Join(dir, "build.sh"), 0755)

	return nil
}

func writeFile(path, content string) {
	_ = os.WriteFile(path, []byte(strings.TrimSpace(content)+"\n"), 0644)
}

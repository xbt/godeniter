// Package godeniter 是一个纯 Go 标准库实现、支持依赖注入、Trie 路由与极简 MVC 的 Web 框架。
package godeniter

import (
	"context"
	"fmt"
	"github.com/xbt/godeniter/inject"
	"github.com/xbt/godeniter/router"
	"github.com/xbt/godeniter/session"
	"html/template"
	"io/fs"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path"
	"path/filepath"
	"reflect"
	"regexp"
	"runtime"
	"strings"
	"syscall"
	"time"

)

// Engine 是 Godeniter 框架的核心实例，负责管理全局依赖注入容器、路由注册与 HTTP 请求分发。
type Engine struct {
	*router.RouterGroup // 组合根路由分组，支持直接在 app 上调用 Get/Post/Use/Group 等方法
	inject.Injector     // 全局依赖注入容器

	router        *router.Router     // 底层 Trie 路由器
	htmlTemplates *template.Template // 全局 HTML 模板实例
	funcMap       template.FuncMap   // 模板渲染函数映射表
	NotFound      HandlerFunc        // 自定义 404 Not Found 处理函数
}

// New 创建并返回一个全新的 Godeniter Engine 实例。
func New() *Engine {
	engine := &Engine{
		Injector: inject.New(),
		router:   router.NewRouter(),
	}
	// 创建根路由分组，并将 Engine 本身作为 bridge 传入
	engine.RouterGroup = router.NewRouterGroup(engine)
	return engine
}

// Classic 创建一个带有常用默认中间件（Logger、Recovery）的 Engine 实例（类似 Martini.Classic()）。
func Classic() *Engine {
	engine := New()
	engine.Use(Logger())
	engine.Use(Recovery())
	return engine
}

// Logger 返回默认请求访问日志中间件。
func Logger() HandlerFunc {
	return func(c *Context) {
		start := time.Now()
		path := c.Req.URL.Path
		raw := c.Req.URL.RawQuery
		if raw != "" {
			path = path + "?" + raw
		}

		c.Next()

		latency := time.Since(start)
		fmt.Printf("[GODENITER] %s | %3d | %13v | %15s | %-7s %s\n",
			time.Now().Format("2006/01/02 - 15:04:05"),
			c.Res.Status(),
			latency,
			c.Req.RemoteAddr,
			c.Req.Method,
			path,
		)
	}
}

// Recovery 返回捕获 Panic 并防止服务器崩溃的恢复中间件。
func Recovery() HandlerFunc {
	return func(c *Context) {
		defer func() {
			if err := recover(); err != nil {
				message := fmt.Sprintf("%s", err)
				stack := traceStack(message)
				fmt.Printf("[GODENITER PANIC RECOVERED]\n%s\n", stack)
				c.String(http.StatusInternalServerError, "500 Internal Server Error\n")
			}
		}()
		c.Next()
	}
}

// Session 返回挂载会话管理的中间件。
func Session(store session.Store, sessionName ...string) HandlerFunc {
	name := "godeniter_session"
	if len(sessionName) > 0 && sessionName[0] != "" {
		name = sessionName[0]
	}

	return func(c *Context) {
		sess, _ := store.Load(c.Req, name)
		if ds, ok := sess.(*session.DefaultSession); ok {
			ds.SetResponseWriter(c.Res)
		}
		c.Set("session", sess)
		c.MapTo(sess, (*session.Session)(nil))

		// 挂载响应头写入前钩子：在发送 WriteHeader (如 c.Redirect 或 c.HTML) 之前先保存 Session Cookie
		if bw, ok := c.Res.(interface{ Before(func()) }); ok {
			bw.Before(func() {
				if sess.IsModified() {
					_ = sess.Save()
				}
			})
		}

		c.Next()

		if sess.IsModified() {
			_ = sess.Save()
		}
	}
}

func traceStack(message string) string {
	var pcs [32]uintptr
	n := runtime.Callers(3, pcs[:])
	var str strings.Builder
	str.WriteString(message + "\nTraceback:")
	for _, pc := range pcs[:n] {
		fn := runtime.FuncForPC(pc)
		file, line := fn.FileLine(pc)
		str.WriteString(fmt.Sprintf("\n\t%s:%d", file, line))
	}
	return str.String()
}

// AddRoute 实现 router.EngineBridge 接口，将路由注册到底层 Router 中。
func (engine *Engine) AddRoute(method string, pattern string, handlers ...interface{}) {
	engine.router.AddRoute(method, pattern, handlers...)
}

// Get 显式注册 GET 路由（解决组合中 Injector.Get 与 RouterGroup.Get 的名称歧义）。
func (engine *Engine) Get(pattern string, handlers ...interface{}) {
	engine.RouterGroup.Get(pattern, handlers...)
}

// GetDependency 从全局注入容器中获取指定类型的依赖对象。
func (engine *Engine) GetDependency(t reflect.Type) reflect.Value {
	return engine.Injector.Get(t)
}

// templateCommentRegex 匹配 <!--{{ ... }}--> 无侵入模板注释语法
var templateCommentRegex = regexp.MustCompile(`<!--\s*\{\{(.*?)\}\}\s*-->`)

// PreprocessHTMLTemplate 将 HTML 模板中的无侵入注释语法 <!--{{ ... }}--> 自动转换为标准 Go 模板语法 {{ ... }}
// 同时完整保留普通开发注释 <!-- ... -->，实现纯前端原型静态无侵入预览与服务端渲染的完美融合。
func PreprocessHTMLTemplate(content string) string {
	return templateCommentRegex.ReplaceAllString(content, "{{$1}}")
}

// ParseHTMLFS 从虚拟文件系统 (fs.FS / embed.FS) 解析 HTML 模板，原生支持 <!--{{ ... }}--> 无侵入注释预处理
func ParseHTMLFS(embedFS fs.FS, funcMap template.FuncMap, patterns ...string) (*template.Template, error) {
	tmpl := template.New("").Funcs(funcMap)
	for _, pattern := range patterns {
		matches, err := fs.Glob(embedFS, pattern)
		if err != nil {
			return nil, err
		}
		for _, match := range matches {
			data, err := fs.ReadFile(embedFS, match)
			if err != nil {
				return nil, err
			}
			preprocessed := PreprocessHTMLTemplate(string(data))
			name := path.Base(match)
			var t *template.Template
			if tmpl.Lookup(name) != nil {
				t = tmpl.Lookup(name)
			} else {
				t = tmpl.New(name)
			}
			if _, err := t.Parse(preprocessed); err != nil {
				return nil, fmt.Errorf("解析模板文件 [%s] 失败: %w", match, err)
			}
		}
	}
	return tmpl, nil
}

// SetFuncMap 为模板引擎注册全局自定义函数字典（请在调用 LoadHTMLFS 或 LoadHTMLGlob 之前设置）。
// 支持流畅链式调用，例如：app.SetFuncMap(template.FuncMap{...}).LoadHTMLFS(...)
func (engine *Engine) SetFuncMap(funcMap template.FuncMap) *Engine {
	engine.funcMap = funcMap
	return engine
}

// SetHTMLTemplate 设置自定义已解析的 HTML 模板对象。
func (engine *Engine) SetHTMLTemplate(tmpl *template.Template) {
	engine.htmlTemplates = tmpl
}

// LoadHTMLGlob 加载匹配 glob 模式的磁盘 HTML 模板文件，自动支持 <!--{{ ... }}--> 无侵入注释预处理。
func (engine *Engine) LoadHTMLGlob(pattern string) {
	matches, err := filepath.Glob(pattern)
	if err != nil {
		log.Fatalf("godeniter: 匹配磁盘模板失败 [%s]: %v", pattern, err)
	}
	tmpl := template.New("").Funcs(engine.funcMap)
	for _, match := range matches {
		data, err := os.ReadFile(match)
		if err != nil {
			log.Fatalf("godeniter: 读取磁盘模板 [%s] 失败: %v", match, err)
		}
		preprocessed := PreprocessHTMLTemplate(string(data))
		name := filepath.Base(match)
		var t *template.Template
		if tmpl.Lookup(name) != nil {
			t = tmpl.Lookup(name)
		} else {
			t = tmpl.New(name)
		}
		if _, err := t.Parse(preprocessed); err != nil {
			log.Fatalf("godeniter: 解析模板文件 [%s] 失败: %v", match, err)
		}
	}
	engine.htmlTemplates = tmpl
}

// LoadHTMLFS 从嵌入式文件系统 (fs.FS / embed.FS) 加载 HTML 模板，为单文件打包提供开箱即用支持。
// 原生支持 <!--{{ ... }}--> 无侵入注释模板预处理，纯前端双击打开预览不破坏布局。
func (engine *Engine) LoadHTMLFS(embedFS fs.FS, patterns ...string) {
	tmpl, err := ParseHTMLFS(embedFS, engine.funcMap, patterns...)
	if err != nil {
		log.Fatalf("godeniter: 加载内嵌模板失败: %v", err)
	}
	engine.htmlTemplates = tmpl
}


// Static 映射磁盘上的静态资源目录到指定的 URL 路由前缀。
// 示例：
//
//	app.Static("/assets", "./public/assets")
func (engine *Engine) Static(relativePath string, root string) {
	handler := engine.createStaticHandler(relativePath, http.Dir(root))
	urlPattern := path.Join(relativePath, "*filepath")
	engine.Get(urlPattern, handler)
	engine.Head(urlPattern, handler)
}

// StaticFS 映射虚拟文件系统（如 embed.FS）中的静态资源到指定的 URL 路由前缀。
// 为 Windows 客户机打包单二进制文件时极为推荐。
func (engine *Engine) StaticFS(relativePath string, fs http.FileSystem) {
	handler := engine.createStaticHandler(relativePath, fs)
	urlPattern := path.Join(relativePath, "*filepath")
	engine.Get(urlPattern, handler)
	engine.Head(urlPattern, handler)
}

func (engine *Engine) createStaticHandler(relativePath string, fs http.FileSystem) HandlerFunc {
	fileServer := http.StripPrefix(relativePath, http.FileServer(fs))
	return func(c *Context) {
		file := c.Param("filepath")
		// 检查静态文件是否存在，若不存在返回 404 状态
		if f, err := fs.Open(file); err != nil {
			c.Status(http.StatusNotFound)
			return
		} else {
			f.Close()
		}
		fileServer.ServeHTTP(c.Res, c.Req)
	}
}

// ServeHTTP 实现 http.Handler 接口，处理接入的每个 HTTP 请求。
func (engine *Engine) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	c := newContext(w, req, engine)

	// 检索匹配的路由节点
	result := engine.router.GetRoute(req.Method, req.URL.Path)
	if result != nil {
		c.Params = result.Params
		c.handlers = result.Handlers
		// 将提取到的动态路由参数映射到当前请求级注入容器中
		c.Map(c.Params)
	} else {
		// 路由未命中：执行自定义或默认 404 处理
		if engine.NotFound != nil {
			c.handlers = []interface{}{engine.NotFound}
		} else {
			c.handlers = []interface{}{func(ctx *Context) {
				ctx.String(http.StatusNotFound, "404 Not Found: %s %s\n", ctx.Method, ctx.Path)
			}}
		}
	}

	// 启动中间件与 Handler 执行链路
	c.Next()
}

// Run 启动 HTTP 监听服务（内置系统中断信号监听与平滑优雅退出 Graceful Shutdown）。
// 特别适合打包为 Windows 可执行文件后，客户双击直接看到访问链接；收到 Ctrl+C 时等待在途请求处理完毕后再安全退出。
func (engine *Engine) Run(addr ...string) error {
	listenAddr := ":8080"
	if len(addr) > 0 && addr[0] != "" {
		listenAddr = addr[0]
	}

	printBanner(listenAddr)

	srv := &http.Server{
		Addr:    listenAddr,
		Handler: engine,
	}

	// 监听中断信号 (Ctrl+C 与 kill 信号)
	stopChan := make(chan os.Signal, 1)
	signal.Notify(stopChan, os.Interrupt, syscall.SIGTERM)

	serverErr := make(chan error, 1)
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			serverErr <- err
		}
	}()

	select {
	case err := <-serverErr:
		fmt.Printf("\n >> [ERROR] Web 服务在 [%s] 启动失败: %v\n", listenAddr, err)
		if strings.Contains(err.Error(), "address already in use") {
			cleanPort := listenAddr
			if cleanPort != "" && cleanPort[0] == ':' {
				cleanPort = cleanPort[1:]
			}
			fmt.Printf(" >> [TIP] 端口 [%s] 已被占用！\n", listenAddr)
			fmt.Printf(" >> [TIP] 解决方案 1: 在终端执行 'lsof -ti :%s | xargs kill -9' 结束占用进程。\n", cleanPort)
			fmt.Printf(" >> [TIP] 解决方案 2: 在 config.json 中修改 \"port\": \":9090\" 使用其他未被占用的端口。\n\n")
		}
		return err
	case sig := <-stopChan:
		fmt.Printf("\n >> 接收到终止信号 [%s]，正在平滑关闭服务 (Graceful Shutdown)...\n", sig)
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := srv.Shutdown(ctx); err != nil {
			fmt.Printf(" >> 强制关闭服务异常: %v\n", err)
			return err
		}
		fmt.Println(" >> 服务已成功安全停止。")
		return nil
	}
}

// printBanner 打印服务启动横幅与本地/局域网访问地址。
func printBanner(addr string) {
	port := addr
	if strings.HasPrefix(addr, ":") {
		port = addr[1:]
	} else if parts := strings.Split(addr, ":"); len(parts) == 2 {
		port = parts[1]
	}

	banner := `
   ______          __            _ __            
  / ____/___  ____/ /__  ____   (_) /____  _____ 
 / / __/ __ \/ __  / _ \/ __ \ / / __/ _ \/ ___/ 
/ /_/ / /_/ / /_/ /  __/ / / // / /_/  __/ /     
\____/\____/\__,_/\___/_/ /_//_/\__/\___/_/      
`
	fmt.Print(banner)
	fmt.Println("==========================================================")
	fmt.Printf(" >> Godeniter 2.0 Web Server Started successfully!\n")
	fmt.Printf(" >> Local URL:    http://127.0.0.1:%s\n", port)
	fmt.Printf(" >> Local URL:    http://localhost:%s\n", port)

	// 获取本机局域网 IP
	if ip := getLocalIP(); ip != "" {
		fmt.Printf(" >> Network URL:  http://%s:%s\n", ip, port)
	}
	fmt.Println("==========================================================")
	fmt.Println(" >> Press Ctrl+C to stop the server.")
}

// getLocalIP 获取本机局域网 IPv4 地址。
func getLocalIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return ""
	}
	for _, address := range addrs {
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}
		}
	}
	return ""
}

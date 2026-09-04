package godeniter

import (
	"encoding/json"
	"fmt"
	"github.com/xbt/godeniter/binding"
	"github.com/xbt/godeniter/inject"
	"github.com/xbt/godeniter/router"
	"github.com/xbt/godeniter/utils/upload"
	"html/template"
	"math"
	"mime/multipart"
	"net/http"
	"reflect"
	"strconv"
)

// abortIndex 定义中断中间件链的最大索引偏移。
const abortIndex int = math.MaxInt8 / 2

// H 是 map[string]interface{} 的快捷类型别名，用于便捷构建 JSON 或模板渲染数据。
type H map[string]interface{}

// HandlerFunc 定义了高性能显式 Context 的标准处理函数签名。
type HandlerFunc func(c *Context)

// Context 封装了单次 HTTP 请求的全部生命周期与环境上下文。
// 内部嵌入了 inject.Injector，具备请求级的依赖注入能力（继承全局 App 注入器）。
type Context struct {
	inject.Injector // 请求级依赖注入容器

	Req      *http.Request  // 原生 HTTP 请求对象
	Res      ResponseWriter // 包装后的 HTTP 响应对象
	Path     string         // 请求的原始 URL Path
	Method   string         // 请求方法 (GET/POST/...)
	Params   router.Params  // 从路由中提取的动态参数 (如 :id, *filepath)
	handlers []interface{}  // 中间件与 Handler 链
	index    int            // 当前执行的中间件索引
	engine   *Engine        // 所属 Engine 指针，用于访问模板等全局资源
	Keys     map[string]any // 上下文元数据暂存字典
}

// newContext 创建并初始化请求上下文实例。
func newContext(w http.ResponseWriter, req *http.Request, engine *Engine) *Context {
	res := newResponseWriter(w)
	inj := inject.New()
	inj.SetParent(engine.Injector) // 继承 Engine 全局注入器

	c := &Context{
		Injector: inj,
		Req:      req,
		Res:      res,
		Path:     req.URL.Path,
		Method:   req.Method,
		Params:   make(router.Params),
		index:    -1,
		engine:   engine,
		Keys:     make(map[string]any),
	}

	// 将 Context 自身以及常用的核心对象注入到请求级容器中
	c.Map(c)
	c.Map(req)
	c.Map(c.Next) // 注入 c.Next 闭包，支持中间件按 func(next func()) 或 c.Next() 调用
	c.MapTo(res, (*ResponseWriter)(nil))
	c.MapTo(res, (*http.ResponseWriter)(nil))

	return c
}

// Next 继续执行中间件链中的下一个处理函数（洋葱圈模型核心）。
func (c *Context) Next() {
	c.index++
	for c.index < len(c.handlers) {
		handler := c.handlers[c.index]
		c.executeHandler(handler)
		c.index++
	}
}

// executeHandler 执行单个中间件或 Handler。
// 既支持极速原生的 func(*Context)，也支持 Martini 风格的任意签名依赖注入函数。
func (c *Context) executeHandler(handler interface{}) {
	if c.IsAborted() {
		return
	}

	switch fn := handler.(type) {
	case HandlerFunc:
		fn(c)
	case func(*Context):
		fn(c)
	case http.HandlerFunc:
		fn(c.Res, c.Req)
	case http.Handler:
		fn.ServeHTTP(c.Res, c.Req)
	default:
		// Martini 风格：通过反射依赖注入执行任意签名的函数
		vals, err := c.Invoke(handler)
		if err != nil {
			panic(fmt.Sprintf("godeniter: 调用 Handler 失败: %v", err))
		}
		// 智能处理返回值（如 return 200, "hello" 或 return H{"msg": "ok"}）
		c.handleReturnValues(vals)
	}
}

// handleReturnValues 智能解析 Martini 风格 Handler 的返回值并写入响应。
func (c *Context) handleReturnValues(vals []reflect.Value) {
	if len(vals) == 0 || c.Res.Written() {
		return
	}

	var status int = http.StatusOK
	var body interface{}

	for _, v := range vals {
		if !v.IsValid() {
			continue
		}
		val := v.Interface()
		switch t := val.(type) {
		case int:
			status = t
		case string:
			body = t
		case []byte:
			body = t
		case error:
			if t != nil {
				c.String(http.StatusInternalServerError, t.Error())
				return
			}
		default:
			body = val
		}
	}

	if body != nil {
		switch b := body.(type) {
		case string:
			c.String(status, b)
		case []byte:
			c.Data(status, "text/plain; charset=utf-8", b)
		default:
			c.JSON(status, b)
		}
	} else if status != http.StatusOK {
		c.Status(status)
	}
}

// Abort 中断中间件链的继续向下执行。
func (c *Context) Abort() {
	c.index = abortIndex
}

// IsAborted 检测当前请求流水线是否已被中断。
func (c *Context) IsAborted() bool {
	return c.index >= abortIndex
}

// AbortWithStatus 写入指定 HTTP 状态码并立即中断后续中间件执行。
func (c *Context) AbortWithStatus(code int) {
	c.Status(code)
	c.Abort()
}

// AbortWithStatusJSON 写入指定状态码与 JSON 数据并立即中断执行。
func (c *Context) AbortWithStatusJSON(code int, jsonObj interface{}) {
	c.Abort()
	c.JSON(code, jsonObj)
}

// Set 在请求生命周期内存储一个键值对。
func (c *Context) Set(key string, val any) {
	c.Keys[key] = val
}

// Get 从请求上下文中获取指定的键值。
func (c *Context) Get(key string) (value any, exists bool) {
	value, exists = c.Keys[key]
	return
}

// Param 获取命名路由参数（例如 "/users/:id" 中的 id）。
func (c *Context) Param(key string) string {
	return c.Params.Get(key)
}

// Query 获取 URL Query 查询参数（例如 "/search?keyword=go" 中的 keyword）。
func (c *Context) Query(key string) string {
	return c.Req.URL.Query().Get(key)
}

// DefaultQuery 获取 URL Query 参数；若不存在则返回默认值 defaultValue。
func (c *Context) DefaultQuery(key string, defaultValue string) string {
	val := c.Query(key)
	if val == "" {
		return defaultValue
	}
	return val
}

// PostForm 获取表单 POST 提交的数据。
func (c *Context) PostForm(key string) string {
	return c.Req.FormValue(key)
}

// DefaultPostForm 获取表单 POST 数据；若不存在则返回默认值 defaultValue。
func (c *Context) DefaultPostForm(key string, defaultValue string) string {
	val := c.PostForm(key)
	if val == "" {
		return defaultValue
	}
	return val
}

// QueryInt 获取整型 URL Query 参数；若不存在或解析失败则返回默认值 defaultValue。
func (c *Context) QueryInt(key string, defaultValue int) int {
	val := c.Query(key)
	if val == "" {
		return defaultValue
	}
	if n, err := strconv.Atoi(val); err == nil {
		return n
	}
	return defaultValue
}

// BindJSON 解析请求体 JSON 并映射到目标结构体。
func (c *Context) BindJSON(obj interface{}) error {
	decoder := json.NewDecoder(c.Req.Body)
	return decoder.Decode(obj)
}

// BindAndValidate 根据 Content-Type 自动解析请求数据并执行 struct tag 校验。
func (c *Context) BindAndValidate(obj interface{}) error {
	return binding.Bind(c.Req, obj)
}

// BindQuery 解析 URL 查询参数并自动校验。
func (c *Context) BindQuery(obj interface{}) error {
	return binding.BindQuery(c.Req, obj)
}

// BindForm 解析 POST/PUT 表单数据并自动校验。
func (c *Context) BindForm(obj interface{}) error {
	return binding.BindForm(c.Req, obj)
}

// FormFile 从 multipart/form-data 表单中获取上传的单文件。
func (c *Context) FormFile(name string) (*multipart.FileHeader, error) {
	if c.Req.MultipartForm == nil {
		if err := c.Req.ParseMultipartForm(32 << 20); err != nil { // 默认最大 32MB 内存缓存
			return nil, err
		}
	}
	f, fh, err := c.Req.FormFile(name)
	if err != nil {
		return nil, err
	}
	f.Close()
	return fh, nil
}

// FormFiles 从 multipart/form-data 表单中获取指定字段名的多个上传文件列表。
func (c *Context) FormFiles(name string) ([]*multipart.FileHeader, error) {
	if c.Req.MultipartForm == nil {
		if err := c.Req.ParseMultipartForm(32 << 20); err != nil {
			return nil, err
		}
	}
	if c.Req.MultipartForm != nil && c.Req.MultipartForm.File != nil {
		if files, ok := c.Req.MultipartForm.File[name]; ok {
			return files, nil
		}
	}
	return nil, http.ErrMissingFile
}

// MultipartForm 获取解析后的完整 MultipartForm 表单。
func (c *Context) MultipartForm() (*multipart.Form, error) {
	err := c.Req.ParseMultipartForm(32 << 20)
	return c.Req.MultipartForm, err
}

// UploadOptions 为 utils/upload.Options 的类型别名，方便直接调用。
type UploadOptions = upload.Options

// SaveUploadedFile 保存上传文件到指定的绝对/相对路径。
func (c *Context) SaveUploadedFile(file *multipart.FileHeader, dst string) error {
	return upload.SaveUploadedFile(file, dst)
}

// SaveUploadedFileWithOptions 按照规则校验并保存上传文件，返回保存后的文件路径。
func (c *Context) SaveUploadedFileWithOptions(file *multipart.FileHeader, opts UploadOptions) (string, error) {
	return upload.SaveUploadedFileWithOptions(file, opts)
}

// Status 设置响应状态码。
func (c *Context) Status(code int) {
	c.Res.WriteHeader(code)
}

// Header 设置响应头。
func (c *Context) Header(key string, value string) {
	c.Res.Header().Set(key, value)
}

// String 向客户端输出格式化的纯文本字符串响应。
func (c *Context) String(code int, format string, values ...interface{}) {
	c.Header("Content-Type", "text/plain; charset=utf-8")
	c.Status(code)
	if len(values) > 0 {
		c.Res.Write([]byte(fmt.Sprintf(format, values...)))
	} else {
		c.Res.Write([]byte(format))
	}
}

// JSON 向客户端输出 JSON 格式响应。
func (c *Context) JSON(code int, obj interface{}) {
	c.Header("Content-Type", "application/json; charset=utf-8")
	c.Status(code)
	encoder := json.NewEncoder(c.Res)
	if err := encoder.Encode(obj); err != nil {
		http.Error(c.Res, err.Error(), http.StatusInternalServerError)
	}
}

// HTML 渲染指定的 HTML 模板并输出响应。
func (c *Context) HTML(code int, name string, data interface{}) {
	c.Header("Content-Type", "text/html; charset=utf-8")
	c.Status(code)
	if c.engine.htmlTemplates != nil {
		if err := c.engine.htmlTemplates.ExecuteTemplate(c.Res, name, data); err != nil {
			c.String(http.StatusInternalServerError, "模板渲染错误: %v", err)
		}
	} else {
		c.String(http.StatusInternalServerError, "未加载任何 HTML 模板")
	}
}

// Data 向客户端输出原始二进制字节数据（如图片、文件下载等）。
func (c *Context) Data(code int, contentType string, data []byte) {
	c.Header("Content-Type", contentType)
	c.Status(code)
	c.Res.Write(data)
}

// APIResponse 定义通用的 REST API 响应结构体。
type APIResponse struct {
	Code    int    `json:"code"`           // 业务状态码 (0 表示成功)
	Message string `json:"message"`        // 提示信息
	Data    any    `json:"data,omitempty"` // 业务数据载荷
}

// Success 以统一的 JSON 格式输出成功响应 (code: 0, message: "ok")。
func (c *Context) Success(data any) {
	c.JSON(http.StatusOK, APIResponse{
		Code:    0,
		Message: "ok",
		Data:    data,
	})
}

// Fail 以统一的 JSON 格式输出业务错误响应。
func (c *Context) Fail(code int, message string) {
	c.JSON(http.StatusOK, APIResponse{
		Code:    code,
		Message: message,
		Data:    nil,
	})
}

// Redirect 执行 HTTP 重定向跳转 (如 301 永久重定向, 302/303 临时重定向)。
func (c *Context) Redirect(code int, location string) {
	http.Redirect(c.Res, c.Req, location, code)
}

// SetCookie 向客户端浏览器写入 HTTP Cookie。
func (c *Context) SetCookie(name, value string, maxAge int, path, domain string, secure, httpOnly bool) {
	if path == "" {
		path = "/"
	}
	http.SetCookie(c.Res, &http.Cookie{
		Name:     name,
		Value:    value,
		MaxAge:   maxAge,
		Path:     path,
		Domain:   domain,
		Secure:   secure,
		HttpOnly: httpOnly,
		SameSite: http.SameSiteLaxMode,
	})
}

// Cookie 获取客户端请求携带的指定 Cookie 值。
func (c *Context) Cookie(name string) (string, error) {
	cookie, err := c.Req.Cookie(name)
	if err != nil {
		return "", err
	}
	return cookie.Value, nil
}

// Session 获取当前请求的 Session 会话对象 (需先挂载 session.Middleware)。
func (c *Context) Session() (any, bool) {
	return c.Get("session")
}

// SetFuncMap 为模板引擎注册自定义函数字典。
func (c *Context) SetFuncMap(funcMap template.FuncMap) {
	c.engine.funcMap = funcMap
}

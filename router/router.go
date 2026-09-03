package router

import (
	"fmt"
	"net/http"
	"strings"
)

// Router 负责管理所有 HTTP 方法的路由树并执行请求路径解析。
type Router struct {
	roots map[string]*node // 每个 HTTP Method (GET/POST/...) 对应一棵独立的 Trie 树
}

// NewRouter 创建并初始化一个空的路由器实例。
func NewRouter() *Router {
	return &Router{
		roots: make(map[string]*node),
	}
}

// AddRoute 向路由器注册一条路由。
// method: HTTP 请求方法 (如 GET, POST, PUT, DELETE, PATCH, OPTIONS, HEAD)
// pattern: 路由路径模式 (如 "/users/:id", "/static/*filepath")
// handlers: 处理函数或中间件列表
func (r *Router) AddRoute(method string, pattern string, handlers ...interface{}) {
	if len(handlers) == 0 {
		panic(fmt.Sprintf("router: 路由 [%s %s] 必须至少指定一个处理函数", method, pattern))
	}

	method = strings.ToUpper(method)
	parts := parsePattern(pattern)

	if _, exists := r.roots[method]; !exists {
		r.roots[method] = &node{}
	}

	r.roots[method].insert(pattern, parts, 0, handlers)
}

// MatchResult 保存路由匹配的结果。
type MatchResult struct {
	Node     *node         // 匹配到的 Trie 节点
	Params   Params        // 解析出的动态路径参数
	Handlers []interface{} // 挂载在该路由上的处理函数与中间件
}

// GetRoute 在对应的 HTTP Method 树中检索匹配路径。
// 返回匹配结果 MatchResult；如果未找到匹配项则返回 nil。
func (r *Router) GetRoute(method string, path string) *MatchResult {
	method = strings.ToUpper(method)
	root, ok := r.roots[method]
	if !ok {
		return nil
	}

	searchParts := parsePattern(path)
	n := root.search(searchParts, 0)
	if n == nil {
		return nil
	}

	// 解析命名动态参数或通配符参数
	params := make(Params)
	registerParts := parsePattern(n.pattern)

	for index, part := range registerParts {
		if part[0] == ':' {
			// 命名参数: ":id" -> params["id"] = searchParts[index]
			if index < len(searchParts) {
				params[part[1:]] = searchParts[index]
			}
		} else if part[0] == '*' && len(part) > 1 {
			// 通配符参数: "*filepath" -> params["filepath"] = "剩余的所有路径"
			if index < len(searchParts) {
				params[part[1:]] = strings.Join(searchParts[index:], "/")
			}
			break
		}
	}

	return &MatchResult{
		Node:     n,
		Params:   params,
		Handlers: n.handlers,
	}
}

// StandardMethods 定义支持的常见 HTTP 方法常量。
var StandardMethods = []string{
	http.MethodGet,
	http.MethodPost,
	http.MethodPut,
	http.MethodDelete,
	http.MethodPatch,
	http.MethodOptions,
	http.MethodHead,
}

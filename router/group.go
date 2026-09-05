// Copyright (c) 2026 xbt. All rights reserved.
// Godeniter is licensed under the GNU General Public License v3.0 (GPL-3.0).
// See LICENSE file in the project root for full license information.

package router

import (
	"path"
)

// IRouter 定义了路由注册及分组的标准接口。
type IRouter interface {
	Use(middlewares ...interface{})
	Group(prefix string, middlewares ...interface{}) *RouterGroup

	Get(pattern string, handlers ...interface{})
	Post(pattern string, handlers ...interface{})
	Put(pattern string, handlers ...interface{})
	Delete(pattern string, handlers ...interface{})
	Patch(pattern string, handlers ...interface{})
	Options(pattern string, handlers ...interface{})
	Head(pattern string, handlers ...interface{})
	Any(pattern string, handlers ...interface{})
}

// RouterGroup 代表一组具有相同路径前缀和公共中间件的路由集合。
type RouterGroup struct {
	prefix      string        // 路由前缀（例如 "/api/v1"）
	middlewares []interface{} // 该分组拥有的公共中间件
	parent      *RouterGroup  // 父级路由分组指针
	engine      EngineBridge  // 底层引擎桥接接口，用于最终将路由注册到底层 Router 中
}

// EngineBridge 定义了 RouterGroup 与上层引擎通信的接口，避免循环引用。
type EngineBridge interface {
	AddRoute(method string, pattern string, handlers ...interface{})
}

// NewRouterGroup 创建一个新的根路由分组。
func NewRouterGroup(engine EngineBridge) *RouterGroup {
	return &RouterGroup{
		prefix:      "",
		middlewares: make([]interface{}, 0),
		engine:      engine,
	}
}

// Group 创建一个子路由分组，继承当前分组的中间件并拼接前缀。
// 示例：
//
//	api := app.Group("/api")
//	v1 := api.Group("/v1", AuthMiddleware())
//	v1.Get("/users", UsersHandler) // 最终路由为: /api/v1/users
func (group *RouterGroup) Group(prefix string, middlewares ...interface{}) *RouterGroup {
	newGroup := &RouterGroup{
		prefix:      group.calculateAbsolutePath(prefix),
		parent:      group,
		engine:      group.engine,
		middlewares: append(group.getCombinedMiddlewares(), middlewares...),
	}
	return newGroup
}

// Use 为当前分组追加一个或多个中间件。
func (group *RouterGroup) Use(middlewares ...interface{}) {
	group.middlewares = append(group.middlewares, middlewares...)
}

// getCombinedMiddlewares 获取当前分组（及所有上级父分组）累加后的全部中间件列表。
func (group *RouterGroup) getCombinedMiddlewares() []interface{} {
	return append([]interface{}{}, group.middlewares...)
}

// calculateAbsolutePath 计算包含所有父级前缀在内的完整绝对路径。
func (group *RouterGroup) calculateAbsolutePath(relativePath string) string {
	if relativePath == "" {
		return group.prefix
	}

	finalPath := path.Join(group.prefix, relativePath)
	// 保持原路径末尾的斜杠特性
	if stringsHasSuffix(relativePath, "/") && !stringsHasSuffix(finalPath, "/") {
		finalPath += "/"
	}
	return finalPath
}

func stringsHasSuffix(s, suffix string) bool {
	return len(s) >= len(suffix) && s[len(s)-len(suffix):] == suffix
}

// handle 组装分组前缀、中间件与目标 Handler 并完成路由注册。
func (group *RouterGroup) handle(method string, relativePath string, handlers ...interface{}) {
	absolutePath := group.calculateAbsolutePath(relativePath)
	// 合并：分组所有中间件 + 当前路由专属 Handlers
	allHandlers := append(group.getCombinedMiddlewares(), handlers...)
	group.engine.AddRoute(method, absolutePath, allHandlers...)
}

// Get 注册一个 GET 请求路由。
func (group *RouterGroup) Get(pattern string, handlers ...interface{}) {
	group.handle("GET", pattern, handlers...)
}

// Post 注册一个 POST 请求路由。
func (group *RouterGroup) Post(pattern string, handlers ...interface{}) {
	group.handle("POST", pattern, handlers...)
}

// Put 注册一个 PUT 请求路由。
func (group *RouterGroup) Put(pattern string, handlers ...interface{}) {
	group.handle("PUT", pattern, handlers...)
}

// Delete 注册一个 DELETE 请求路由。
func (group *RouterGroup) Delete(pattern string, handlers ...interface{}) {
	group.handle("DELETE", pattern, handlers...)
}

// Patch 注册一个 PATCH 请求路由。
func (group *RouterGroup) Patch(pattern string, handlers ...interface{}) {
	group.handle("PATCH", pattern, handlers...)
}

// Options 注册一个 OPTIONS 请求路由。
func (group *RouterGroup) Options(pattern string, handlers ...interface{}) {
	group.handle("OPTIONS", pattern, handlers...)
}

// Head 注册一个 HEAD 请求路由。
func (group *RouterGroup) Head(pattern string, handlers ...interface{}) {
	group.handle("HEAD", pattern, handlers...)
}

// Any 注册匹配所有标准 HTTP 方法的路由 (GET, POST, PUT, DELETE, PATCH, OPTIONS, HEAD)。
func (group *RouterGroup) Any(pattern string, handlers ...interface{}) {
	for _, method := range StandardMethods {
		group.handle(method, pattern, handlers...)
	}
}

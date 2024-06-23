package main

import (
	"godeniter/middleware"
	"net/http"
)

type route struct {
	pattern string
	handler http.HandlerFunc
}

var routes = []route{}

func addRoute(pattern string, handler http.HandlerFunc) {
	routes = append(routes, route{pattern, handler})
}

func router() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		for _, route := range routes {
			if r.URL.Path == route.pattern {
				handler := applyMiddlewares(http.HandlerFunc(route.handler))
				handler.ServeHTTP(w, r)
				return
			}
		}
		http.NotFound(w, r)
	})
}

type Middleware interface {
	Handle(next http.Handler) http.Handler
}

var middlewares []Middleware

func use(mw Middleware) {
	middlewares = append(middlewares, mw)
}

func applyMiddlewares(handler http.Handler) http.Handler {
	for i := len(middlewares) - 1; i >= 0; i-- {
		handler = middlewares[i].Handle(handler)
	}
	return handler
}

func init() {
	// 使用访问控制中间件
	use(middleware.NewAccessControlMiddleware())
	// 使用日志中间件
	use(middleware.NewLoggingMiddleware())
}

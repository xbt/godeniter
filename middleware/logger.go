// Package middleware 提供了 Godeniter 框架的常用内置中间件。
package middleware

import (
	"fmt"
	"github.com/xbt/godeniter/router"
	"net/http"
	"time"
)

// ResponseStatusProvider 定义了能够获取状态码的接口。
type ResponseStatusProvider interface {
	Status() int
}

// Logger 返回一个格式化请求日志中间件。
// 打印内容包括：请求时间、HTTP 状态码、请求处理耗时、客户端 IP、HTTP 动词与请求路径。
func Logger() func(res http.ResponseWriter, req *http.Request, next func()) {
	return func(res http.ResponseWriter, req *http.Request, next func()) {
		start := time.Now()
		path := req.URL.Path
		raw := req.URL.RawQuery
		if raw != "" {
			path = path + "?" + raw
		}

		// 执行后续处理流水线
		next()

		latency := time.Since(start)
		statusCode := http.StatusOK
		if sp, ok := res.(ResponseStatusProvider); ok {
			statusCode = sp.Status()
		}

		// 根据状态码输出格式
		fmt.Printf("[GODENITER] %s | %3d | %13v | %15s | %-7s %s\n",
			time.Now().Format("2006/01/02 - 15:04:05"),
			statusCode,
			latency,
			req.RemoteAddr,
			req.Method,
			path,
		)
	}
}

// ParamsFromContext 辅助提取路由参数（若有）。
func ParamsFromContext(params router.Params, key string) string {
	if params == nil {
		return ""
	}
	return params.Get(key)
}

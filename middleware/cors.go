package middleware

import (
	"net/http"
)

// CORSOptions 定义跨域资源共享配置项。
type CORSOptions struct {
	AllowOrigins     []string // 允许的 Origin 来源
	AllowMethods     []string // 允许的 HTTP 方法
	AllowHeaders     []string // 允许的 Header 标头
	ExposeHeaders    []string // 允许前端获取的 Header
	AllowCredentials bool     // 是否允许跨域携带凭证 (Cookie)
}

// CORS 返回一个跨域资源共享处理中间件。
func CORS(opts ...CORSOptions) func(res http.ResponseWriter, req *http.Request, next func()) {
	var opt CORSOptions
	if len(opts) > 0 {
		opt = opts[0]
	} else {
		// 默认允许全开放跨域
		opt = CORSOptions{
			AllowOrigins: []string{"*"},
			AllowMethods: []string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS", "HEAD"},
			AllowHeaders: []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Requested-With"},
		}
	}

	return func(res http.ResponseWriter, req *http.Request, next func()) {
		origin := req.Header.Get("Origin")
		if origin != "" {
			res.Header().Set("Access-Control-Allow-Origin", origin)
		} else if len(opt.AllowOrigins) > 0 {
			res.Header().Set("Access-Control-Allow-Origin", opt.AllowOrigins[0])
		}

		if len(opt.AllowMethods) > 0 {
			res.Header().Set("Access-Control-Allow-Methods", joinStrings(opt.AllowMethods, ", "))
		}
		if len(opt.AllowHeaders) > 0 {
			res.Header().Set("Access-Control-Allow-Headers", joinStrings(opt.AllowHeaders, ", "))
		}
		if opt.AllowCredentials {
			res.Header().Set("Access-Control-Allow-Credentials", "true")
		}

		// 处理预检请求 (OPTIONS)
		if req.Method == http.MethodOptions {
			res.WriteHeader(http.StatusNoContent)
			return
		}

		next()
	}
}

func joinStrings(elems []string, sep string) string {
	switch len(elems) {
	case 0:
		return ""
	case 1:
		return elems[0]
	}
	var b []byte
	for i, s := range elems {
		if i > 0 {
			b = append(b, sep...)
		}
		b = append(b, s...)
	}
	return string(b)
}

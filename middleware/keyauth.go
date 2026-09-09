// Copyright (c) 2026 xbt. All rights reserved.
// Godeniter is licensed under the GNU General Public License v3.0 (GPL-3.0).
// See LICENSE file in the project root for full license information.

package middleware

import (
	"net/http"
	"strings"
)

// KeyAuthOptions API Key 鉴权中间件配置
type KeyAuthOptions struct {
	QueryParam   string                                          // Query 参数名称，默认 "api_key"
	HeaderKey    string                                          // 自定义请求头名称，默认 "X-API-Key"
	ErrorHandler func(res http.ResponseWriter, req *http.Request) // 自定义鉴权失败回调
}

// KeyAuth 返回一个轻量纯 Go 的 API Key / Bearer Token 鉴权中间件。
// 优先从 Authorization: Bearer <token> 提取，次从 X-API-Key 头提取，最后从 Query 参数提取。
func KeyAuth(validator func(key string) bool, opts ...KeyAuthOptions) func(res http.ResponseWriter, req *http.Request, next func()) {
	opt := KeyAuthOptions{
		QueryParam: "api_key",
		HeaderKey:  "X-API-Key",
	}
	if len(opts) > 0 {
		if opts[0].QueryParam != "" {
			opt.QueryParam = opts[0].QueryParam
		}
		if opts[0].HeaderKey != "" {
			opt.HeaderKey = opts[0].HeaderKey
		}
		opt.ErrorHandler = opts[0].ErrorHandler
	}

	return func(res http.ResponseWriter, req *http.Request, next func()) {
		var token string

		// 1. Authorization: Bearer <token>
		authHeader := req.Header.Get("Authorization")
		if strings.HasPrefix(authHeader, "Bearer ") {
			token = strings.TrimPrefix(authHeader, "Bearer ")
		} else if strings.HasPrefix(authHeader, "bearer ") {
			token = strings.TrimPrefix(authHeader, "bearer ")
		}

		// 2. 自定义请求头 (如 X-API-Key)
		if token == "" && opt.HeaderKey != "" {
			token = req.Header.Get(opt.HeaderKey)
		}

		// 3. Query 参数
		if token == "" && opt.QueryParam != "" {
			token = req.URL.Query().Get(opt.QueryParam)
		}

		// 4. 校验
		if token == "" || (validator != nil && !validator(token)) {
			if opt.ErrorHandler != nil {
				opt.ErrorHandler(res, req)
			} else {
				res.Header().Set("Content-Type", "application/json; charset=utf-8")
				res.WriteHeader(http.StatusUnauthorized)
				_, _ = res.Write([]byte(`{"code":401,"message":"401 Unauthorized: Invalid or missing API Key"}`))
			}
			return
		}

		next()
	}
}

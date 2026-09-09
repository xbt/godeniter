// Copyright (c) 2026 xbt. All rights reserved.
// Godeniter is licensed under the GNU General Public License v3.0 (GPL-3.0).
// See LICENSE file in the project root for full license information.

package middleware

import (
	"net/http"
	"strings"
)

// SecurityOptions Web 安全头配置选项
type SecurityOptions struct {
	XContentTypeOptions string // 默认 "nosniff"
	XFrameOptions       string // 默认 "SAMEORIGIN"
	XXSSProtection      string // 默认 "1; mode=block"
	ReferrerPolicy      string // 默认 "strict-origin-when-cross-origin"
}

// Security 返回标准 Web 安全防护头中间件。
// 自动为所有 HTTP 响应注入 X-Content-Type-Options、X-Frame-Options、X-XSS-Protection、Referrer-Policy 标头。
func Security(opts ...SecurityOptions) func(res http.ResponseWriter, req *http.Request, next func()) {
	opt := SecurityOptions{
		XContentTypeOptions: "nosniff",
		XFrameOptions:       "SAMEORIGIN",
		XXSSProtection:      "1; mode=block",
		ReferrerPolicy:      "strict-origin-when-cross-origin",
	}
	if len(opts) > 0 {
		if opts[0].XContentTypeOptions != "" {
			opt.XContentTypeOptions = opts[0].XContentTypeOptions
		}
		if opts[0].XFrameOptions != "" {
			opt.XFrameOptions = opts[0].XFrameOptions
		}
		if opts[0].XXSSProtection != "" {
			opt.XXSSProtection = opts[0].XXSSProtection
		}
		if opts[0].ReferrerPolicy != "" {
			opt.ReferrerPolicy = opts[0].ReferrerPolicy
		}
	}

	return func(res http.ResponseWriter, req *http.Request, next func()) {
		if opt.XContentTypeOptions != "" {
			res.Header().Set("X-Content-Type-Options", opt.XContentTypeOptions)
		}
		if opt.XFrameOptions != "" {
			res.Header().Set("X-Frame-Options", opt.XFrameOptions)
		}
		if opt.XXSSProtection != "" {
			res.Header().Set("X-XSS-Protection", opt.XXSSProtection)
		}
		if opt.ReferrerPolicy != "" {
			res.Header().Set("Referrer-Policy", opt.ReferrerPolicy)
		}

		next()
	}
}

// DefaultSensitivePatterns 默认拦截的敏感文件/路径特征
var DefaultSensitivePatterns = []string{
	".db", ".sqlite", ".sqlite3", ".sql",
	"/data/", "config.json", ".env", ".log", ".pid",
	".git", ".svn", ".htaccess",
}

// BlockSensitive 返回敏感资源探测拦截中间件。
// 自动拦截外部对 SQLite 库、配置文件、环境变量、版本库等敏感文件的探测请求，直接返回 403 Forbidden。
func BlockSensitive(customPatterns ...string) func(res http.ResponseWriter, req *http.Request, next func()) {
	patterns := make([]string, 0, len(DefaultSensitivePatterns)+len(customPatterns))
	patterns = append(patterns, DefaultSensitivePatterns...)
	patterns = append(patterns, customPatterns...)

	return func(res http.ResponseWriter, req *http.Request, next func()) {
		reqPath := strings.ToLower(req.URL.Path)
		for _, pattern := range patterns {
			if strings.Contains(reqPath, strings.ToLower(pattern)) {
				res.Header().Set("Content-Type", "application/json; charset=utf-8")
				res.WriteHeader(http.StatusForbidden)
				_, _ = res.Write([]byte(`{"code":403,"message":"403 Forbidden: Access to protected resource is denied."}`))
				return
			}
		}

		next()
	}
}

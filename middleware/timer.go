// Copyright (c) 2026 xbt. All rights reserved.
// Godeniter is licensed under the GNU General Public License v3.0 (GPL-3.0).
// See LICENSE file in the project root for full license information.

package middleware

import (
	"fmt"
	"net/http"
	"time"
)

// ServerTiming 返回请求耗时与 W3C Server-Timing 标头注入中间件。
// 为响应注入 X-Response-Time 与 Server-Timing 标头，便于前端与链路追踪工具分析后端接口耗时。
func ServerTiming() func(res http.ResponseWriter, req *http.Request, next func()) {
	return func(res http.ResponseWriter, req *http.Request, next func()) {
		start := time.Now()

		next()

		duration := time.Since(start)
		durMs := fmt.Sprintf("%.2fms", float64(duration.Microseconds())/1000.0)

		res.Header().Set("X-Response-Time", durMs)
		res.Header().Set("Server-Timing", fmt.Sprintf("app;dur=%.2f", float64(duration.Microseconds())/1000.0))
	}
}

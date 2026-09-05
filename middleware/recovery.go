// Copyright (c) 2026 xbt. All rights reserved.
// Godeniter is licensed under the GNU General Public License v3.0 (GPL-3.0).
// See LICENSE file in the project root for full license information.

package middleware

import (
	"fmt"
	"net/http"
	"runtime"
	"strings"
)

// Recovery 返回一个防止服务崩溃的 Panic 捕获中间件。
// 当下游 Handler 发生 panic 时，该中间件会捕获异常、在控制台打印详细堆栈调用信息，并向客户端返回 500 Internal Server Error。
func Recovery() func(res http.ResponseWriter, req *http.Request, next func()) {
	return func(res http.ResponseWriter, req *http.Request, next func()) {
		defer func() {
			if err := recover(); err != nil {
				message := fmt.Sprintf("%s", err)
				stack := trace(message)
				fmt.Printf("[GODENITER PANIC RECOVERED]\n%s\n", stack)

				http.Error(res, "Internal Server Error", http.StatusInternalServerError)
			}
		}()

		next()
	}
}

// trace 获取发生 panic 时的堆栈跟踪信息。
func trace(message string) string {
	var pcs [32]uintptr
	n := runtime.Callers(3, pcs[:]) // 跳过前 3 层调用栈

	var str strings.Builder
	str.WriteString(message + "\nTraceback:")
	for _, pc := range pcs[:n] {
		fn := runtime.FuncForPC(pc)
		file, line := fn.FileLine(pc)
		str.WriteString(fmt.Sprintf("\n\t%s:%d", file, line))
	}
	return str.String()
}

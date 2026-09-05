// Copyright (c) 2026 xbt. All rights reserved.
// Godeniter is licensed under the GNU General Public License v3.0 (GPL-3.0).
// See LICENSE file in the project root for full license information.

package godeniter

import (
	"bufio"
	"net"
	"net/http"
)

// ResponseWriter 包装了原生的 http.ResponseWriter，
// 增加了对 HTTP 状态码记录、响应体字节数统计以及写入状态检测的能力。
type ResponseWriter interface {
	http.ResponseWriter
	http.Hijacker
	http.Flusher

	// Status 返回当前响应已写入的 HTTP 状态码。若尚未写入，默认返回 200。
	Status() int

	// Size 返回已写入响应体的总字节数。
	Size() int

	// Written 返回响应头是否已经被刷新写入。
	Written() bool

	// Before 注册一个在响应头实际写入网络前执行的钩子函数 (如用于写入 Session Cookie 等)。
	Before(func())
}

// responseWriter 是 ResponseWriter 接口的默认实现。
type responseWriter struct {
	http.ResponseWriter
	status      int
	size        int
	written     bool
	beforeHooks []func()
}

// newResponseWriter 创建并初始化一个 ResponseWriter 包装器。
func newResponseWriter(w http.ResponseWriter) *responseWriter {
	return &responseWriter{
		ResponseWriter: w,
		status:         http.StatusOK, // 默认为 200 OK
		size:           0,
		written:        false,
		beforeHooks:    make([]func(), 0),
	}
}

// Before 注册一个在 WriteHeader 真正执行前触发的回调钩子。
func (w *responseWriter) Before(fn func()) {
	if fn != nil {
		w.beforeHooks = append(w.beforeHooks, fn)
	}
}

func (w *responseWriter) triggerBefore() {
	for _, fn := range w.beforeHooks {
		fn()
	}
	w.beforeHooks = nil
}

// WriteHeader 记录状态码并写入底层响应头。
func (w *responseWriter) WriteHeader(code int) {
	if !w.written {
		w.triggerBefore()
		w.status = code
		w.written = true
		w.ResponseWriter.WriteHeader(code)
	}
}

// Write 写入响应数据体，并统计写入字节数。
func (w *responseWriter) Write(data []byte) (int, error) {
	if !w.written {
		w.WriteHeader(http.StatusOK)
	}
	n, err := w.ResponseWriter.Write(data)
	w.size += n
	return n, err
}

// Status 获取 HTTP 状态码。
func (w *responseWriter) Status() int {
	return w.status
}

// Size 获取已写入的响应体长度。
func (w *responseWriter) Size() int {
	return w.size
}

// Written 获取是否已经输出响应头。
func (w *responseWriter) Written() bool {
	return w.written
}

// Flush 实现 http.Flusher 接口。
func (w *responseWriter) Flush() {
	if flusher, ok := w.ResponseWriter.(http.Flusher); ok {
		flusher.Flush()
	}
}

// Hijack 实现 http.Hijacker 接口（支持 WebSocket 等协议升级场景）。
func (w *responseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	if hijacker, ok := w.ResponseWriter.(http.Hijacker); ok {
		return hijacker.Hijack()
	}
	return nil, nil, http.ErrNotSupported
}

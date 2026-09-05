// Copyright (c) 2026 xbt. All rights reserved.
// Godeniter is licensed under the GNU General Public License v3.0 (GPL-3.0).
// See LICENSE file in the project root for full license information.

package session

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

// CookieStore 基于加密签名防篡改 Cookie 的会话存储器（最适合无状态单文件交付）。
type CookieStore struct {
	secretKey []byte // 用于 HMAC 签名的密钥
	maxAge    int    // Cookie 有效期（秒），默认 86400 (1天)
	path      string // Cookie Path，默认 "/"
}

// NewCookieStore 创建一个 CookieStore 实例。secretKey 为签名私钥。
func NewCookieStore(secretKey string) *CookieStore {
	return &CookieStore{
		secretKey: []byte(secretKey),
		maxAge:    86400 * 7, // 默认 7 天
		path:      "/",
	}
}

// Load 从请求 Cookie 中读取会话并校验 HMAC 签名。
func (cs *CookieStore) Load(r *http.Request, name string) (Session, error) {
	s := &DefaultSession{
		values:   make(map[string]any),
		modified: false,
		store:    cs,
		name:     name,
	}

	cookie, err := r.Cookie(name)
	if err != nil || cookie.Value == "" {
		return s, nil
	}

	// Cookie 格式: base64(payload).signature
	parts := strings.SplitN(cookie.Value, ".", 2)
	if len(parts) != 2 {
		return s, fmt.Errorf("session: 非法的 cookie 格式")
	}

	payloadBase64 := parts[0]
	signature := parts[1]

	// 校验 HMAC 签名是否匹配
	expectedSig := cs.sign(payloadBase64)
	if !hmac.Equal([]byte(signature), []byte(expectedSig)) {
		return s, fmt.Errorf("session: cookie 签名校验失败 (数据可能已被篡改)")
	}

	payloadBytes, err := base64.URLEncoding.DecodeString(payloadBase64)
	if err != nil {
		return s, err
	}

	var values map[string]any
	if err := json.Unmarshal(payloadBytes, &values); err != nil {
		return s, err
	}

	s.values = values
	return s, nil
}

// Save 将会话数据序列化并写回 HTTP 响应头。
func (cs *CookieStore) Save(w http.ResponseWriter, name string, s Session) error {
	ds, ok := s.(*DefaultSession)
	if !ok {
		return fmt.Errorf("session: 必须是 DefaultSession 实例")
	}

	ds.mu.RLock()
	payloadBytes, err := json.Marshal(ds.values)
	ds.mu.RUnlock()
	if err != nil {
		return err
	}

	payloadBase64 := base64.URLEncoding.EncodeToString(payloadBytes)
	signature := cs.sign(payloadBase64)
	cookieValue := payloadBase64 + "." + signature

	http.SetCookie(w, &http.Cookie{
		Name:     name,
		Value:    cookieValue,
		MaxAge:   cs.maxAge,
		Path:     cs.path,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})

	return nil
}

// sign 生成 HMAC-SHA256 十六进制签名。
func (cs *CookieStore) sign(data string) string {
	h := hmac.New(sha256.New, cs.secretKey)
	h.Write([]byte(data))
	return hex.EncodeToString(h.Sum(nil))
}

// ContextBinder 定义支持数据暂存与依赖注入的通用上下文接口。
type ContextBinder interface {
	Set(key string, val any)
	MapTo(val any, ifacePtr any) any
}

// Middleware 返回一个全自动加载与保存 Session 的中间件。
// 示例：
//
//	store := session.NewCookieStore("my-secret-key-123456")
//	app.Use(session.Middleware(store, "godeniter_session"))
func Middleware(store Store, sessionName string) func(res http.ResponseWriter, req *http.Request, next func()) {
	if sessionName == "" {
		sessionName = "godeniter_session"
	}

	return func(res http.ResponseWriter, req *http.Request, next func()) {
		sess, _ := store.Load(req, sessionName)
		if ds, ok := sess.(*DefaultSession); ok {
			ds.w = res
		}

		// 执行业务链路
		next()

		// 请求处理完毕后，如果有修改则自动保存写回响应
		if sess.IsModified() {
			_ = sess.Save()
		}
	}
}

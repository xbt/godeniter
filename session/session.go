// Package session 提供了类似 PHP $_SESSION 体验的纯 Go 标准库服务端会话管理模块。
// 支持 HMAC-SHA256 安全签名的 CookieStore 以及内存 Store，支持开箱即用的中间件无缝注入。
package session

import (
	"net/http"
	"sync"
)

// Session 定义了会话操作的标准接口。
type Session interface {
	// Get 获取指定键的值
	Get(key string) any
	// GetString 获取字符串类型的值，若不存在或类型不匹配返回空字符串
	GetString(key string) string
	// GetInt 获取整型值，若不存在返回 0
	GetInt(key string) int
	// Set 设置键值对
	Set(key string, val any)
	// Delete 删除指定键
	Delete(key string)
	// Clear 清空所有会话数据
	Clear()
	// SetFlash 设置一次性闪存消息 (FlashData)，类似 CodeIgniter $this->session->set_flashdata()
	// 存储的数据在下一次被读取 (GetFlash) 后自动销毁，非常适合表单提交后的重定向提示
	SetFlash(key string, val any)
	// GetFlash 读取并销毁指定的一次性闪存消息
	GetFlash(key string) any
	// GetFlashString 读取并销毁指定的一次性闪存消息字符串
	GetFlashString(key string) string
	// Save 将会话数据序列化并保存（写回 Cookie / 存储器）
	Save() error
	// IsModified 检查会话是否被修改过
	IsModified() bool
}

// Store 定义了会话存储器的接口规范。
type Store interface {
	// Load 从 HTTP 请求中提取并加载 Session
	Load(r *http.Request, name string) (Session, error)
	// Save 将 Session 保存写回 HTTP 响应中
	Save(w http.ResponseWriter, name string, s Session) error
}

// DefaultSession 是 Session 接口的基础实现。
type DefaultSession struct {
	mu       sync.RWMutex
	values   map[string]any
	modified bool
	store    Store
	w        http.ResponseWriter
	name     string
}

// NewSession 创建一个新的空会话。
func NewSession(store Store, w http.ResponseWriter, name string) *DefaultSession {
	return &DefaultSession{
		values:   make(map[string]any),
		modified: false,
		store:    store,
		w:        w,
		name:     name,
	}
}

// SetResponseWriter 设置写回 Cookie 的 HTTP 响应对象。
func (s *DefaultSession) SetResponseWriter(w http.ResponseWriter) {
	s.w = w
}

func (s *DefaultSession) Get(key string) any {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.values[key]
}

func (s *DefaultSession) GetString(key string) string {
	val := s.Get(key)
	if str, ok := val.(string); ok {
		return str
	}
	return ""
}

func (s *DefaultSession) GetInt(key string) int {
	val := s.Get(key)
	switch v := val.(type) {
	case int:
		return v
	case float64:
		return int(v)
	case int64:
		return int(v)
	default:
		return 0
	}
}

func (s *DefaultSession) Set(key string, val any) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.values[key] = val
	s.modified = true
}

func (s *DefaultSession) Delete(key string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.values, key)
	s.modified = true
}

func (s *DefaultSession) Clear() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.values = make(map[string]any)
	s.modified = true
}

func (s *DefaultSession) SetFlash(key string, val any) {
	s.Set("_flash_"+key, val)
}

func (s *DefaultSession) GetFlash(key string) any {
	flashKey := "_flash_" + key
	s.mu.Lock()
	defer s.mu.Unlock()
	val, ok := s.values[flashKey]
	if ok {
		delete(s.values, flashKey)
		s.modified = true
	}
	return val
}

func (s *DefaultSession) GetFlashString(key string) string {
	val := s.GetFlash(key)
	if str, ok := val.(string); ok {
		return str
	}
	return ""
}

func (s *DefaultSession) IsModified() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.modified
}

func (s *DefaultSession) Save() error {
	if s.store != nil && s.w != nil {
		err := s.store.Save(s.w, s.name, s)
		if err == nil {
			s.mu.Lock()
			s.modified = false
			s.mu.Unlock()
		}
		return err
	}
	return nil
}

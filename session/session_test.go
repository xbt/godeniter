package session

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCookieStore_SaveAndLoad(t *testing.T) {
	secret := "test-secret-key-12345"
	store := NewCookieStore(secret)

	// 1. 模拟写入 Session
	w := httptest.NewRecorder()
	sess := NewSession(store, w, "test_session")
	sess.Set("user_id", 1001)
	sess.Set("username", "ben")

	if err := sess.Save(); err != nil {
		t.Fatalf("Session.Save 失败: %v", err)
	}

	cookie := w.Result().Cookies()[0]
	if cookie.Name != "test_session" || cookie.Value == "" {
		t.Fatalf("Cookie 未正确写入: %+v", cookie)
	}

	// 2. 模拟携带 Cookie 发起后续请求并读取
	req := httptest.NewRequest("GET", "/profile", nil)
	req.AddCookie(cookie)

	loadedSess, err := store.Load(req, "test_session")
	if err != nil {
		t.Fatalf("store.Load 失败: %v", err)
	}

	if loadedSess.GetString("username") != "ben" {
		t.Errorf("GetString(username) 错误: %s", loadedSess.GetString("username"))
	}
	if loadedSess.GetInt("user_id") != 1001 {
		t.Errorf("GetInt(user_id) 错误: %d", loadedSess.GetInt("user_id"))
	}

	// 3. 模拟篡改 Cookie 测试防伪验证
	tamperedCookie := &http.Cookie{
		Name:  "test_session",
		Value: "tampered_data." + cookie.Value,
	}
	tamperedReq := httptest.NewRequest("GET", "/profile", nil)
	tamperedReq.AddCookie(tamperedCookie)

	_, err = store.Load(tamperedReq, "test_session")
	if err == nil {
		t.Errorf("被篡改的 Cookie 应当校验失败")
	}
}

func TestSession_FlashData(t *testing.T) {
	sess := NewSession(nil, nil, "flash_test")

	// 设置 Flash 消息
	sess.SetFlash("notice", "文章创建成功")

	// 第一次读取：应该正常获取
	msg := sess.GetFlashString("notice")
	if msg != "文章创建成功" {
		t.Errorf("期望获取到 Flash 消息 '文章创建成功'，实际获取: %s", msg)
	}

	// 第二次读取：应该已被自动销毁，返回空
	msgSecond := sess.GetFlashString("notice")
	if msgSecond != "" {
		t.Errorf("Flash 消息读取后应当自动销毁，但仍获取到: %s", msgSecond)
	}
}

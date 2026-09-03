package godeniter

import (
	"encoding/json"
	"godeniter/router"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type AppConfig struct {
	Env string
}

func TestEngine_BasicAndDI(t *testing.T) {
	app := New()
	cfg := &AppConfig{Env: "production"}
	app.Map(cfg)

	// 1. 标准 Context Handler
	app.Get("/ping", func(c *Context) {
		c.JSON(http.StatusOK, H{"message": "pong"})
	})

	// 2. Martini 风格 Handler (参数依赖注入 + 返回值自动渲染)
	app.Get("/users/:id", func(params router.Params, config *AppConfig) (int, H) {
		return http.StatusOK, H{
			"user_id": params.Get("id"),
			"env":     config.Env,
		}
	})

	// 测试 /ping
	req := httptest.NewRequest("GET", "/ping", nil)
	w := httptest.NewRecorder()
	app.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("期望状态码 200，收到: %d", w.Code)
	}
	var res map[string]string
	json.Unmarshal(w.Body.Bytes(), &res)
	if res["message"] != "pong" {
		t.Errorf("响应内容不符合预期: %v", res)
	}

	// 测试 /users/:id 依赖注入
	req = httptest.NewRequest("GET", "/users/999", nil)
	w = httptest.NewRecorder()
	app.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("期望状态码 200，收到: %d", w.Code)
	}
	var userRes map[string]string
	json.Unmarshal(w.Body.Bytes(), &userRes)
	if userRes["user_id"] != "999" || userRes["env"] != "production" {
		t.Errorf("DI 返回值与参数不符合预期: %v", userRes)
	}
}

func TestEngine_MiddlewareOnion(t *testing.T) {
	app := New()
	calls := make([]string, 0)

	// 中间件 1
	app.Use(func(c *Context) {
		calls = append(calls, "mw1_start")
		c.Next()
		calls = append(calls, "mw1_end")
	})

	// 中间件 2
	app.Use(func(c *Context) {
		calls = append(calls, "mw2_start")
		c.Next()
		calls = append(calls, "mw2_end")
	})

	app.Get("/test", func(c *Context) {
		calls = append(calls, "handler")
		c.String(http.StatusOK, "ok")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	app.ServeHTTP(w, req)

	expectedCalls := []string{"mw1_start", "mw2_start", "handler", "mw2_end", "mw1_end"}
	if len(calls) != len(expectedCalls) {
		t.Fatalf("中间件执行链路长度不匹配: %v", calls)
	}
	for i, name := range expectedCalls {
		if calls[i] != name {
			t.Errorf("中间件执行顺序错误，期望 [%s]，实际 [%s]", name, calls[i])
		}
	}
}

func TestEngine_Recovery(t *testing.T) {
	app := New()
	app.Use(Recovery())

	app.Get("/panic", func(c *Context) {
		panic("something went wrong!")
	})

	req := httptest.NewRequest("GET", "/panic", nil)
	w := httptest.NewRecorder()
	app.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Recovery 中间件未正确返回 500 状态码，收到: %d", w.Code)
	}
}

func TestEngine_GroupAndPost(t *testing.T) {
	app := New()
	api := app.Group("/api/v1")

	type LoginPayload struct {
		Username string `json:"username"`
	}

	api.Post("/login", func(c *Context) {
		var p LoginPayload
		if err := c.BindJSON(&p); err != nil {
			c.String(http.StatusBadRequest, "invalid json")
			return
		}
		c.JSON(http.StatusOK, H{"token": "token_for_" + p.Username})
	})

	body := strings.NewReader(`{"username": "ben"}`)
	req := httptest.NewRequest("POST", "/api/v1/login", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	app.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("POST 绑定响应状态码错误: %d", w.Code)
	}
	var res map[string]string
	json.Unmarshal(w.Body.Bytes(), &res)
	if res["token"] != "token_for_ben" {
		t.Errorf("Token 返回值不符合预期: %v", res)
	}
}

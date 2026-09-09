// Copyright (c) 2026 xbt. All rights reserved.
// Godeniter is licensed under the GNU General Public License v3.0 (GPL-3.0).
// See LICENSE file in the project root for full license information.

package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSecurity(t *testing.T) {
	mw := Security()
	req := httptest.NewRequest("GET", "/", nil)
	rec := httptest.NewRecorder()

	calledNext := false
	mw(rec, req, func() {
		calledNext = true
	})

	if !calledNext {
		t.Fatal("expected next() to be called")
	}
	if rec.Header().Get("X-Content-Type-Options") != "nosniff" {
		t.Errorf("expected nosniff, got %s", rec.Header().Get("X-Content-Type-Options"))
	}
	if rec.Header().Get("X-Frame-Options") != "SAMEORIGIN" {
		t.Errorf("expected SAMEORIGIN, got %s", rec.Header().Get("X-Frame-Options"))
	}
}

func TestBlockSensitive(t *testing.T) {
	mw := BlockSensitive()

	// 1. 探测敏感文件
	req := httptest.NewRequest("GET", "/data/app.db", nil)
	rec := httptest.NewRecorder()
	calledNext := false
	mw(rec, req, func() {
		calledNext = true
	})
	if calledNext {
		t.Fatal("expected request to be blocked for /data/app.db")
	}
	if rec.Code != http.StatusForbidden {
		t.Errorf("expected status 403, got %d", rec.Code)
	}

	// 2. 正常请求
	reqOK := httptest.NewRequest("GET", "/api/articles", nil)
	recOK := httptest.NewRecorder()
	calledNextOK := false
	mw(recOK, reqOK, func() {
		calledNextOK = true
	})
	if !calledNextOK {
		t.Fatal("expected request to pass for /api/articles")
	}
}

func TestServerTiming(t *testing.T) {
	mw := ServerTiming()
	req := httptest.NewRequest("GET", "/test", nil)
	rec := httptest.NewRecorder()

	called := false
	mw(rec, req, func() {
		called = true
	})

	if !called {
		t.Fatal("expected next() to be called")
	}
	if rec.Header().Get("X-Response-Time") == "" {
		t.Errorf("expected X-Response-Time header")
	}
	if rec.Header().Get("Server-Timing") == "" {
		t.Errorf("expected Server-Timing header")
	}
}

func TestKeyAuth(t *testing.T) {
	mw := KeyAuth(func(key string) bool {
		return key == "secret-token-123"
	})

	// 1. 无 token
	reqNoToken := httptest.NewRequest("GET", "/api/v1", nil)
	recNoToken := httptest.NewRecorder()
	called := false
	mw(recNoToken, reqNoToken, func() { called = true })
	if called || recNoToken.Code != http.StatusUnauthorized {
		t.Fatal("expected 401 without token")
	}

	// 2. Bearer Token
	reqBearer := httptest.NewRequest("GET", "/api/v1", nil)
	reqBearer.Header.Set("Authorization", "Bearer secret-token-123")
	recBearer := httptest.NewRecorder()
	called = false
	mw(recBearer, reqBearer, func() { called = true })
	if !called {
		t.Fatal("expected pass with valid bearer token")
	}

	// 3. X-API-Key
	reqAPIKey := httptest.NewRequest("GET", "/api/v1", nil)
	reqAPIKey.Header.Set("X-API-Key", "secret-token-123")
	recAPIKey := httptest.NewRecorder()
	called = false
	mw(recAPIKey, reqAPIKey, func() { called = true })
	if !called {
		t.Fatal("expected pass with valid X-API-Key header")
	}

	// 4. Query param
	reqQuery := httptest.NewRequest("GET", "/api/v1?api_key=secret-token-123", nil)
	recQuery := httptest.NewRecorder()
	called = false
	mw(recQuery, reqQuery, func() { called = true })
	if !called {
		t.Fatal("expected pass with valid api_key query")
	}
}

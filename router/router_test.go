package router

import (
	"reflect"
	"testing"
)

type mockEngine struct {
	routes map[string]string // "METHOD /path" -> handler summary
}

func (m *mockEngine) AddRoute(method string, pattern string, handlers ...interface{}) {
	m.routes[method+" "+pattern] = pattern
}

func newTestRouter() *Router {
	r := NewRouter()
	r.AddRoute("GET", "/", "rootHandler")
	r.AddRoute("GET", "/hello/:name", "helloHandler")
	r.AddRoute("GET", "/users/:id/posts/:post_id", "userPostHandler")
	r.AddRoute("GET", "/static/*filepath", "staticHandler")
	r.AddRoute("POST", "/users", "createUserHandler")
	return r
}

func TestRouter_ParsePattern(t *testing.T) {
	ok := reflect.DeepEqual(parsePattern("/p/:name"), []string{"p", ":name"})
	ok = ok && reflect.DeepEqual(parsePattern("/p/*"), []string{"p", "*"})
	ok = ok && reflect.DeepEqual(parsePattern("/p/*name/*"), []string{"p", "*name"})
	if !ok {
		t.Fatal("parsePattern 解析路由模式失败")
	}
}

func TestRouter_GetRoute(t *testing.T) {
	r := newTestRouter()

	// 1. 测试根路径
	res := r.GetRoute("GET", "/")
	if res == nil || res.Node.pattern != "/" {
		t.Errorf("根路径匹配失败")
	}

	// 2. 测试单动态参数
	res = r.GetRoute("GET", "/hello/godeniter")
	if res == nil || res.Node.pattern != "/hello/:name" {
		t.Fatalf("单参数路由匹配失败")
	}
	if res.Params.Get("name") != "godeniter" {
		t.Errorf("动态参数解析错误，期望 'godeniter'，得到: %s", res.Params.Get("name"))
	}

	// 3. 测试多动态参数
	res = r.GetRoute("GET", "/users/100/posts/200")
	if res == nil || res.Node.pattern != "/users/:id/posts/:post_id" {
		t.Fatalf("多参数路由匹配失败")
	}
	if res.Params.Get("id") != "100" || res.Params.Get("post_id") != "200" {
		t.Errorf("多参数解析错误: id=%s, post_id=%s", res.Params.Get("id"), res.Params.Get("post_id"))
	}

	// 4. 测试通配符全路径
	res = r.GetRoute("GET", "/static/css/style.css")
	if res == nil || res.Node.pattern != "/static/*filepath" {
		t.Fatalf("通配符路由匹配失败")
	}
	if res.Params.Get("filepath") != "css/style.css" {
		t.Errorf("通配符路径捕获错误，期望 'css/style.css'，得到: %s", res.Params.Get("filepath"))
	}

	// 5. 测试 POST 方法
	res = r.GetRoute("POST", "/users")
	if res == nil || res.Node.pattern != "/users" {
		t.Errorf("POST 路由匹配失败")
	}

	// 6. 测试未定义路由 (404)
	res = r.GetRoute("GET", "/not/found")
	if res != nil {
		t.Errorf("不应匹配不存在的路径")
	}
}

func TestRouterGroup(t *testing.T) {
	mockEng := &mockEngine{routes: make(map[string]string)}
	rootGroup := NewRouterGroup(mockEng)

	mw1 := "globalMiddleware"
	mw2 := "apiMiddleware"
	mw3 := "v1Middleware"

	rootGroup.Use(mw1)

	api := rootGroup.Group("/api", mw2)
	v1 := api.Group("/v1", mw3)

	v1.Get("/users", "v1UsersHandler")
	v1.Post("/users", "v1CreateUserHandler")

	if _, ok := mockEng.routes["GET /api/v1/users"]; !ok {
		t.Errorf("分组路由前缀拼接错误: GET /api/v1/users 缺失")
	}
	if _, ok := mockEng.routes["POST /api/v1/users"]; !ok {
		t.Errorf("分组路由前缀拼接错误: POST /api/v1/users 缺失")
	}
}

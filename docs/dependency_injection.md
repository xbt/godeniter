# 依赖注入系统手册 (Dependency Injection)

Godeniter 内置了基于反射实现的轻量级依赖注入容器（`inject`），借鉴了 **Martini** 的设计哲学。

---

## 1. 核心原理

在传统的 Go Web 框架中，Handler 的签名必须严格为 `func(c *Context)` 或 `func(w http.ResponseWriter, r *http.Request)`。
而在 Godeniter 中，Handler 的参数签名是**完全自由**的，您可以声明任何已注册到容器中的依赖类型（如配置、数据库句柄、业务 Service、Session 等）：

```go
app.Get("/users", func(svc *UserService, db *db.DB) (int, godeniter.H) {
    users := svc.GetAll()
    return 200, godeniter.H{"data": users}
})
```

---

## 2. 注册依赖项

### (1) 按具体类型映射 (`Map`)
```go
type AppConfig struct {
    Env  string
    Port int
}

cfg := &AppConfig{Env: "production", Port: 8080}
app.Map(cfg) // 注入 *AppConfig 类型到全局容器
```

### (2) 按接口类型映射 (`MapTo`)
```go
type Logger interface {
    Log(msg string)
}

type FileLogger struct{}
func (f *FileLogger) Log(msg string) { /* ... */ }

app.MapTo(&FileLogger{}, (*Logger)(nil)) // 将具体结构体注入到 Logger 接口类型
```

---

## 3. 在 Handler 中自动注入

```go
app.Get("/info", func(cfg *AppConfig, logger Logger) string {
    logger.Log("Accessing /info")
    return "App is running in: " + cfg.Env
})
```

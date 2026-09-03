# 模式一：前后端分离 (RESTful API + SPA) 快速开发示例

本示例展示了如何使用 **Godeniter 2.0** 构建现代化前后端分离的 Web 应用与 RESTful API 服务。

---

## 🌟 核心特性与设计

1. **统一 API 响应契约**：
   * 成功响应：`c.Success(data)` 输出 `{"code": 0, "message": "ok", "data": ...}`
   * 失败响应：`c.Fail(40001, "错误原因")`
2. **依赖注入业务服务**：
   * 将业务 Service（如 `ArticleService`）通过 `app.Map(service)` 注入全局容器，Controller 中直接作为参数声明即可接收。
3. **CORS 跨域支持**：
   * 框架内置 `middleware.CORS()`，一键解决前后端分离跨域预检与携带凭证问题。
4. **单文件 SPA 内嵌交付**：
   * 通过 `//go:embed static/*` 将前端 HTML/JS 页面内嵌至单一二进制文件，后端同时具备纯 API 能力与开箱即用的前端交互界面。

---

## 🛠️ 接口清单

| 方法 | 路由 | 说明 | 示例请求/响应 |
| :--- | :--- | :--- | :--- |
| `GET` | `/` | SPA 交互管理面板 | 返回前端单页应用 |
| `GET` | `/api/v1/articles` | 获取文章列表 | `{"code":0,"data":{"total":2,"items":[...]}}` |
| `GET` | `/api/v1/articles/:id`| 获取文章详情 | `{"code":0,"data":{"id":1,"title":"..."}}` |
| `POST` | `/api/v1/articles` | 新增文章 (JSON Body) | `{"title":"...","content":"...","author":"..."}` |
| `DELETE` | `/api/v1/articles/:id`| 删除文章 | `{"code":0,"data":{"deleted_id":1}}` |

---

## 🚀 运行与构建

### 1. 本地运行调试
```bash
go run ./examples/01_api_spa/main.go
```
启动后在浏览器打开：`http://127.0.0.1:8080`

### 2. 编译为 Windows 单文件 (.exe)
```bash
CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o dist/api-spa-app.exe ./examples/01_api_spa/main.go
```
将生成的 `api-spa-app.exe` 交付给客户，双击即可直接使用！

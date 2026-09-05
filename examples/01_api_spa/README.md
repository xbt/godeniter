# 模式一：前后端分离 (RESTful API + SPA 单页) 完整项目示例

本示例展示了如何使用 **Godeniter 2.0** 构建现代化前后端分离的 Web 应用，包含 **文件上传**、**数据分页**、**模糊搜索**、**参数校验** 与 **敏感信息脱敏**。

---

## 🌟 核心功能与架构亮点

1. **统一 API 响应规范**：
   * 成功响应：`c.Success(data)` 输出 `{"code": 0, "message": "ok", "data": ...}`
   * 失败响应：`c.Fail(40001, "错误原因")`
2. **基于 Struct Tag 的参数绑定与校验 (`binding/`)**：
   * 在 Controller 中直接使用 `c.BindAndValidate(&req)`，自动执行 `required,min=2,max=50` 等规则。
3. **文件上传与安全存储 (`UploadOptions`)**：
   * 支持本地图片上传（`POST /api/v1/upload`），限制文件大小（5MB）、扩展名白名单（`.jpg`, `.png` 等），自动生成时间戳+随机防重名安全文件名。
4. **数据库与内存分页检索 (`Paginate`)**：
   * 接口支持 `GET /api/v1/articles?page=1&page_size=5&keyword=golang`，自动输出分页元数据 `total`、`page`、`total_pages`、`has_next`、`has_prev`。
5. **阅读量快捷自增与脱敏展示 (`utils/str`)**：
   * 访问详情时阅读量自动 `Increment`。
   * 作者手机号/邮箱通过 `str.MaskPhone` / `str.MaskEmail` 自动脱敏为 `138****5678`。
6. **单文件 SPA 内嵌交付**：
   * 通过 `//go:embed static/*` 将前端现代化管理面板打包进单一 `.exe`，双击运行即直接可用！

---

## 🛠️ API 接口列表

| 方法 | 路由 | 说明 | 示例参数 / 响应 |
| :--- | :--- | :--- | :--- |
| `GET` | `/` | SPA 交互管理面板 | 返回前端单页应用（支持上传、搜索与分页） |
| `GET` | `/api/v1/articles` | 分页与模糊搜索文章列表 | Query: `?page=1&page_size=5&keyword=go` |
| `GET` | `/api/v1/articles/:id`| 获取文章详情 (自动自增阅读量) | 返回文章详情与脱敏作者信息 |
| `POST` | `/api/v1/articles` | 新增文章 (带参数校验) | JSON Body: `{"title":"...","author":"...","content":"..."}` |
| `POST` | `/api/v1/upload` | 上传封面图片 | Multipart Form: `file` -> 返回 `{"url": "/uploads/images/..."}` |
| `DELETE` | `/api/v1/articles/:id`| 删除文章 | `{"code":0,"data":{"deleted_id":1}}` |

---

## 🚀 运行与单文件打包

### 1. 本地运行调试
```bash
go run ./examples/01_api_spa/main.go
```
启动后在浏览器中打开：`http://127.0.0.1:8080`

### 2. 编译为 Windows 单文件可执行程序 (.exe)
```bash
CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o dist/demo1_api_spa.exe ./examples/01_api_spa/main.go
```
将生成的 `demo1_api_spa.exe` 发送给客户，在 Windows 机器上双击即可运行！

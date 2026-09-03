# 模式二：经典 PHP 风格服务端渲染 (MVC + HTML Template) 完整项目示例

本示例展示了如何使用 **Godeniter 2.0** 构建类似 PHP CodeIgniter / Laravel 风格的服务端渲染（SSR）Web 应用，包含 **Session 登录认证**、**头像文件上传**、**工单多条件模糊搜索** 与 **字符串脱敏/摘要截断**。

---

## 🌟 核心特性与架构亮点

1. **经典 MVC 分层架构**：
   * **Controllers (`controllers/`)**：处理 HTTP 请求、调用服务、通过 `c.HTML(200, "index.html", data)` 渲染视图。
   * **Models (`models/`)**：实体数据模型，配合 `db.Table(...)` 链式查询操作。
   * **Views (`views/`)**：Go 原生 `html/template`，支持数据注入、循环列表与条件分支渲染。
2. **服务端 Session 会话管理 (`session/`)**：
   * 基于 HMAC-SHA256 安全签名的防篡改 CookieStore，Controller 中直接声明 `sess session.Session` 即可使用 `sess.Set(...)` 与 `sess.GetString(...)`。
3. **用户头像文件上传 (`c.FormFile`)**：
   * 支持通过传统 HTML Form POST 表单上传用户头像，校验文件大小与扩展名（`.jpg`, `.png`），保存并即时在界面回显。
4. **字符串辅助处理 (`strutil`)**：
   * 使用 `strutil.UUID()` 和 `MD5` 自动生成格式化工单编号（`PRJ-xxxx`）。
   * 使用 `strutil.Truncate` 对长文本进行智能截断与摘要提取。
   * 使用 `strutil.MaskPhone` 对手机号码进行脱敏保护。
5. **单文件内嵌打包**：
   * 通过 `//go:embed views/*` 将所有 HTML 模板自动打入单个可执行程序，部署到客户机器上绝无“模板丢失”烦恼。

---

## 🚀 运行与单文件打包

### 1. 本地运行调试
```bash
go run ./examples/02_mvc_template/main.go
```
启动后在浏览器中打开：`http://127.0.0.1:8080`
* 默认管理员账号：`admin`
* 默认管理员密码：`123456`
* 登录后可测试头像上传、工单关键词搜索与安全注销。

### 2. 编译为 Windows 单文件可执行程序 (.exe)
```bash
CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o dist/demo2_mvc_template.exe ./examples/02_mvc_template/main.go
```
将生成的 `demo2_mvc_template.exe` 直接发送给客户，双击即可直接运行！

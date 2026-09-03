# 模式二：经典 PHP 风格服务端渲染 (MVC + HTML Template) 开发示例

本示例展示了如何使用 **Godeniter 2.0** 构建类似 PHP CodeIgniter / Laravel Blade 风格的服务端渲染（SSR）Web 应用。

---

## 🌟 核心特性与设计

1. **经典 MVC 分层架构**：
   * **Controllers (`controllers/`)**：负责获取请求、调用模型或服务、决定渲染哪个视图。
   * **Models (`models/`)**：数据实体定义，结合 `db.Table(...)` 开展数据库读写。
   * **Views (`views/`)**：Go 原生 `html/template`，支持数据注入、条件分支与循环渲染。
2. **零前端构建工具链 (Zero Build Step)**：
   * 不需要 Node.js、Webpack、Vite 或 npm，直接编写 HTML/CSS 即可运行。
3. **Session / Cookie 会话管理**：
   * 通过 `c.SetCookie(...)` 与 `c.Cookie(...)` 轻松实现登录态维护与注销。
4. **单文件内嵌打包**：
   * 通过 `//go:embed views/*` 将所有 HTML 模板自动打入二进制文件，避免部署时“模板文件路径丢失”的问题。

---

## 🚀 运行与构建

### 1. 本地运行调试
```bash
go run ./examples/02_mvc_template/main.go
```
启动后在浏览器打开：`http://127.0.0.1:8080`，可测试浏览列表、登录（账号 `admin`，密码 `123456`）和注销流程。

### 2. 编译为 Windows 单文件 (.exe)
```bash
CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o dist/mvc-template-app.exe ./examples/02_mvc_template/main.go
```
生成独立的 `mvc-template-app.exe`，直接发给客户双击即可运行！

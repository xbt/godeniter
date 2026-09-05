# 跨平台单文件打包与 Windows 客户机交付手册 (Build & Deploy)

Godeniter 支持将所有前端静态页面、HTML 模板和 Go 后端程序打包为一个单一的可执行文件（Windows 下为 `.exe`），客户机双击即可运行，无需预装任何环境。

---

## 1. 原生内嵌原理 (Go embed)

```go
package main

import (
    "embed"
    "github.com/xbt/godeniter"
    "net/http"
)

//go:embed static/*
var staticFS embed.FS

func main() {
    app := godeniter.Classic()

    // 读取内嵌静态单页
    app.Get("/", func(c *godeniter.Context) {
        data, _ := staticFS.ReadFile("static/index.html")
        c.Data(http.StatusOK, "text/html; charset=utf-8", data)
    })

    app.Run(":8080")
}
```

---

## 2. 跨平台一键编译命令

### (1) Windows 64位统一全能单文件编译
```bash
# Windows 统一单文件 (0 外部依赖，免 CGO 交叉编译)
CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o dist/app.exe .

# macOS / Linux 统一二进制
go build -ldflags="-s -w" -o dist/app .
```

### (2) 纯 Go 标准库 Windows 图标编译器 (`utils/rsrc`)
项目根目录下只要放置 `app.ico` 图标：
```bash
# 动态检测 app.ico 并自动编译生成 resource_windows_amd64.syso (0 外部依赖，纯标准库)
go run github.com/xbt/godeniter/cmd/rsrc -auto
```
后续执行 `go build` 时，Go 工具链会自动将该图标资源段缝合进 `.exe` 二进制中，在 Windows 资源管理器中呈现专属桌面图标，同时自动作为网页 `/favicon.ico` 响应！

### (3) 使用官方一键打包脚本
```bash
./build.sh     # macOS / Linux
build.bat      # Windows
```

---

## 3. 客户机交付与使用体验

将编译生成的 `dist/app.exe` 拷贝至 Windows 客户机器：
1. **终端 CLI 运维**：在 CMD / PowerShell 运行 `app.exe start/status/stop/restart`，全生命周期指令清晰友好，不跳多余黑框；
2. **桌面双击 / 托盘模式**：双击 `app.exe`（或运行 `app.exe tray`），底层自动调用 Win32 原生 API 隐去控制台黑框，右下角直接点亮专属托盘图标，右键提供“打开管理后台、打开应用目录、关于系统、退出程序”原生菜单，双击图标直接调起默认浏览器，交付体验极佳！

---

## 4. UPX 极速瘦身与 Windows 数字签名 (防拦截)

若需要使用 UPX 进一步压缩可执行文件体积，或消除 Windows SmartScreen 蓝底拦截警告与杀软误报，请参阅：
👉 **[《Windows 数字签名与代码证书实战手册 (docs/code_signing.md)》](./code_signing.md)**：
* 包含纯 Go 标准库生成自签证书 (`cmd/cert`)；
* UPX 压缩与数字签名的正确流水线顺序；
* 客户机一键导入公钥 (`certutil`) 终身免告警实战。

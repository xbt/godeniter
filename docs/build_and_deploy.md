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

### (1) macOS / Linux 编译 Windows 64位独立程序
```bash
# 1. Windows 控制台版 (双击运行弹出控制台窗口，方便查看实时请求日志)
CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o dist/app.exe .

# 2. Windows 纯静默托盘客户端 (彻底隐藏控制台黑框，双击后直接常驻右下角通知区域)
CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -ldflags="-s -w -H=windowsgui" -o dist/app_tray.exe .

# 3. macOS / Linux 统一二进制
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

## 3. 客户机交付效果

将编译生成的 `dist/` 文件拷贝至客户机器：
1. **控制台版 `app.exe`**：双击弹出终端，显示 ASCII 启动横幅并提示本机访问地址（如 `http://127.0.0.1:8080`），适合服务器环境；
2. **托盘版 `app_tray.exe`**：双击**无任何黑框**，右下角直接点亮专属托盘图标，右键提供“打开管理后台、打开应用目录、关于系统、退出程序”原生菜单，双击图标直接调起默认浏览器，交付体验极佳！

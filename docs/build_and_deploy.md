# 跨平台单文件打包与 Windows 客户机交付手册 (Build & Deploy)

Godeniter 支持将所有前端静态页面、HTML 模板和 Go 后端程序打包为一个单一的可执行文件（Windows 下为 `.exe`），客户机双击即可运行，无需预装任何环境。

---

## 1. 原生内嵌原理 (Go embed)

```go
package main

import (
    "embed"
    "godeniter"
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

### (1) macOS / Linux 编译 Windows 64位独立 `.exe`
```bash
CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o dist/my_app.exe ./main.go
```

### (2) 使用官方一键打包脚本
```bash
./build.sh     # macOS / Linux
build.bat      # Windows
```

---

## 3. 客户机运行效果

将编译生成的 `dist/my_app.exe` 拷贝至客户机器：
1. 客户直接**双击** `my_app.exe`。
2. 终端自动弹出 ASCII 启动横幅并提示本机访问地址（如 `http://127.0.0.1:8080`）。
3. 客户在浏览器中打开即可使用完整系统。

# 跨平台系统托盘与状态栏常驻模式开发手册 (`tray/`)

**Godeniter** 原生内置了专为 Web 服务打造的轻量级、跨平台**系统托盘与菜单栏模式（Tray & Status Bar）**。
能够让开发者基于同一个代码库，既能在 Linux/云端作为守护进程运行，也能在个人电脑（macOS / Windows）上作为常驻系统的桌面客户端分发。

---

## 🎯 核心架构与 0 外部依赖设计

我们严格贯彻框架 **0 第三方依赖** 的核心哲学，在多平台下均以最高原生度实现：

1. **Windows 平台（100% 纯 Go 标准库，0 CGO）**：
   * 位于 `tray/tray_windows.go`；
   * 通过标准库 `syscall.NewLazyDLL` 动态链接 Windows 原生 `user32.dll` 与 `shell32.dll`；
   * **零 CGO、零第三方包**：在 macOS / Linux 上执行 `CGO_ENABLED=0 GOOS=windows go build` 交叉编译秒级通过，绝对零编译门槛；
   * **原生 ICO 动态提取**：在内存中原生解析 `.ico` 二进制字节流，自动匹配最适合的 16x16 / 32x32 尺寸并调用 `CreateIconFromResourceEx` 创建系统 `HICON`；
   * **Win32 规范防失灵**：严格遵循微软官方 KB139526 规范，在弹出菜单关闭时向窗口投递 `WM_NULL`，彻底避免点击外部收起后二次右键失灵的问题。

2. **macOS 平台（系统内置 Cocoa / AppKit 驱动）**：
   * 位于 `tray/tray_darwin.go`、`tray/tray_darwin.h`、`tray/tray_darwin.m`；
   * 仅调用 macOS 出厂自带的 AppKit 框架（`NSStatusBar` / `NSStatusItem` / `NSMenu`），无任何第三方 npm 或 Electron 依赖；
   * **Accessory 常驻模式**：配置为 `NSApplicationActivationPolicyAccessory`，仅在屏幕右上角状态栏常驻，**不占用底部 Dock 栏**；
   * **自适应宽度与图文并茂**：使用 `NSVariableStatusItemLength` 自适应宽度，采用 `[图标] 标题` 模式，无论在深色模式还是带刘海（Notch）的 MacBook Pro 屏幕上均醒目可见；
   * **退出无感清理**：在退出循环前显式移除状态栏 Item，实现瞬间无感退出。

3. **Linux / 无 GUI 环境（优雅降级）**：
   * 位于 `tray/tray_other.go`；
   * 自动降级为标准控制台信号监听模式，接收到中断信号时安全平滑关闭。

---

## 📋 原生菜单四件套与交互规范

点击 macOS 状态栏图标或右键 Windows 托盘图标，默认呈现经过最佳实践检验的经典四件套：

| 菜单项 | 图标 | 交互行为 |
| :--- | :---: | :--- |
| **打开管理后台** | 🌐 | 跨平台调起系统默认浏览器访问 Web 服务主页（如 `http://127.0.0.1:8080`） |
| **打开应用目录** | 📁 | 调起系统文件管理器（Windows 资源管理器 / Mac Finder）定位到应用运行目录，方便查找 `config.json`、`data/app.db` 与 `app.log` |
| **关于系统** | ℹ️ | 弹出原生系统关于弹窗，展示应用名、版本、PID、端口及 Go 运行时环境 |
| ────────── | ── | 原生视觉分割线 |
| **退出程序** | ⏹️ | 触发 `OnExit` 优雅停机钩子，平滑释放 HTTP 连接，销毁托盘图标并安全退出 |

*同时支持通过 `Options.Menus` 声明任意数量自定义业务菜单。*

---

## 💻 快速调用示例

```go
package main

import (
    "context"
    "fmt"
    "net/http"
    "time"

    "github.com/xbt/godeniter"
    "github.com/xbt/godeniter/tray"
)

//go:embed app.ico
var appIcoBytes []byte

func main() {
    app := godeniter.Classic()
    app.Get("/", func(c *godeniter.Context) {
        c.String(200, "Hello Godeniter Desktop!")
    })

    port := ":8080"
    webURL := "http://127.0.0.1" + port

    // 1. 异步协程启动 Web 服务
    srv := &http.Server{
        Addr:    port,
        Handler: app,
    }
    go func() {
        if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            fmt.Printf("Web 服务异常退出: %v\n", err)
        }
    }()

    // 2. 主线程运行跨平台系统托盘
    _ = tray.Run(tray.Options{
        Title:     "Godeniter Desktop",
        Tooltip:   "Godeniter 桌面客户端 (" + port + ")",
        IconBytes: appIcoBytes, // 自动适配 Windows 托盘与 macOS 状态栏
        URL:       webURL,
        Version:   "v1.0.0",
        Port:      port,
        OnExit: func() {
            fmt.Println("正在安全平滑关闭服务...")
            ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
            defer cancel()
            _ = srv.Shutdown(ctx)
        },
    })
}
```

---

## 📦 跨平台打包与无黑框静默客户端

### 1. 打包命令

```bash
# 生成标准控制台版 (双击运行保留控制台窗口，方便查看调试日志)
CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o dist/app.exe .

# 生成纯静默桌面托盘客户端 (彻底隐藏控制台黑框，直接在托盘常驻！)
CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -ldflags="-s -w -H=windowsgui" -o dist/app_tray.exe .

# 生成 macOS / Linux 统一二进制
go build -ldflags="-s -w" -o dist/app .
```

### 2. 脚手架一键构建
在 `godeniter-starter` 脚手架中，直接执行：
```bash
./build.sh     # macOS / Linux
build.bat      # Windows
```
脚本将自动检测 `app.ico` 动态编译资源段，并同时在 `dist/` 目录下生成上述三种交付物！

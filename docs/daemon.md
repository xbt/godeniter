# Godeniter 服务生命周期与守护进程运维手册 (Daemon & Lifecycle)

Godeniter 坚持 **100% 纯 Go 标准库（0 外部第三方依赖）** 设计，在核心库中原生内置了轻量级、工业级的**服务生命周期与守护进程管理器 (`godeniter/daemon`)**。

无论是通过源码直接运行（`go run main.go <cmd>`），还是交付打包后的可执行单文件（`./app <cmd>`），开发者与运维人员均可享受类似 Nginx / PM2 的守护进程管理体验。

---

## 🌟 核心特性与优势

1. **0 外部依赖**：仅依托 Go 原生 `os`, `os/exec`, `os/signal`, `syscall` 实现，没有引入任何庞杂的 CGO 或第三方库。
2. **双模自由切换**：
   - **前台开发模式**：默认无参数启动时，前台阻塞运行，输出彩色 ASCII Banner、实时打印访问日志，支持随时按 `Ctrl+C` 平滑退出。
   - **后台守护模式**：输入 `start` 或在配置中开启 `"daemon": true`，进程自动 fork 脱离控制终端（Setsid），主命令输出 PID 后**立即返回命令行**，日志重定向到文件，断开 SSH / 关闭终端完全不影响服务运行。
3. **跨平台原生适配**：
   - **Linux / macOS / Unix**：使用 `syscall.Setsid` 剥离终端控制会话，通过 `syscall.SIGTERM` 触发框架优雅停机，通过 `syscall.Signal(0)` 实现轻量存活性探测。
   - **Windows**：通过 `CreationFlags: DETACHED_PROCESS` 剥离控制台窗口，确保跨平台交叉编译交付 `.exe` 时 **0 编译报错**。
4. **统一指令规范**：提供标准的 `start`、`stop`、`restart`、`status` 指令集。

---

## 🚀 快速上手与指令指南

### 1. 常用生命周期指令速查

无论是在本地开发环境，还是在 Linux 服务器上，指令完全一致：

| 指令 | 源码执行方式 | 二进制执行方式 | 作用与说明 |
| :--- | :--- | :--- | :--- |
| **默认前台运行** | `go run main.go` | `./app` | 阻塞在当前终端，实时打印访问日志与彩色 Banner，适合本地日常开发调试。按 `Ctrl+C` 停止。 |
| **后台启动 (守护模式)** | `go run main.go start` | `./app start` | 自动在后台拉起子进程，脱离当前终端，将 PID 写入 `app.pid`，日志重定向至 `app.log`，**立即返回命令行**。 |
| **查看运行状态** | `go run main.go status` | `./app status` | 读取 `app.pid`，自动探测后台进程是否存活，显示运行状态、PID 及日志文件位置。若异常退出自动清理残留 PID。 |
| **平滑安全停止** | `go run main.go stop` | `./app stop` | 向后台进程发送 `SIGTERM` 信号，触发框架内置优雅停机（释放连接池与挂起任务），确认退出后自动清理 PID 文件。 |
| **平滑重启服务** | `go run main.go restart` | `./app restart` | 自动先执行 `stop`，等待旧进程完全退出释放端口后，重新执行 `start` 拉起新进程。 |

---

## 💻 运维终端操作实录

### 1. 后台启动服务
```bash
$ go run main.go start
>> [CONFIG] 成功从本地加载配置文件: config.json
==========================================================
 >> Godeniter 2.0 服务已成功在后台启动 (Daemon Mode)!
 >> 进程 PID:    81547 (已写入 ./app.pid)
 >> 监听端口:    :8080
 >> 输出日志:    ./app.log
----------------------------------------------------------
 >> 运维常用指令 (源码运行 / 二进制运行):
    - 查看状态: go run main.go status  (或 ./app status)
    - 停止服务: go run main.go stop    (或 ./app stop)
    - 重启服务: go run main.go restart (或 ./app restart)
    - 实时日志: tail -f ./app.log
==========================================================
$ 
# 终端直接返回！关闭窗口、断开 SSH 完全不中断服务。

```

### 2. 查看服务状态与日志
```bash
# 查看状态
$ go run main.go status
==========================================================
 >> Godeniter 服务状态: [运行中 🟢]
 >> 运行 PID:    81547
 >> PID 文件:    ./app.pid
 >> 日志文件:    ./app.log
==========================================================

# 实时追踪日志 (动态滚屏，按 Ctrl+C 仅退出查看，不影响服务)
$ tail -f ./app.log
```

### 3. 重启与停止服务
```bash
# 平滑重启
$ go run main.go restart
>> [CONFIG] 成功从本地加载配置文件: config.json
>> [RESTART] 正在平滑重启服务...
>> [STOP] 正在向服务发送安全退出信号 (PID: 81547)...
>> [STOP] Godeniter 服务已安全优雅退出！
==========================================================
 >> Godeniter 2.0 服务已成功在后台启动 (Daemon Mode)!
 >> 进程 PID:    81580 (已写入 ./app.pid)
 ...

# 停止服务
$ go run main.go stop
>> [CONFIG] 成功从本地加载配置文件: config.json
>> [STOP] 正在向服务发送安全退出信号 (PID: 81580)...
>> [STOP] Godeniter 服务已安全优雅退出！

# 再次查看确认
$ go run main.go status
>> [STATUS] Godeniter 服务状态: [未运行] (PID 文件不存在)
```

---

## ⚙️ 配置文件驱动 (`config.json`)

除了命令行参数外，系统支持通过 `config.json` 声明式启用守护进程：

```json
{
  "app": {
    "name": "Godeniter Application",
    "port": ":8080",
    "daemon": true,             // 设为 true 时，直接执行 go run main.go 或 ./app 也会自动进入后台运行
    "pid_file": "./data/app.pid", // 自定义 PID 存储路径
    "log_file": "./logs/app.log"  // 自定义日志存储路径
  }
}
```

* **环境变量覆盖**：支持通过 `APP_DAEMON=true`、`PID_FILE=/var/run/app.pid`、`LOG_FILE=/var/log/app.log` 覆盖配置。

---

## 🏗️ 核心接入代码 (3 行极简集成)

在基于 Godeniter 开发的任意项目中，仅需 3 行代码即可将服务委托给生命周期管理器：

```go
package main

import (
    "github.com/xbt/godeniter"
    "github.com/xbt/godeniter/daemon"
)

func main() {
    app := godeniter.Classic()
    
    // 注册你的路由与中间件
    app.Get("/", func(c *godeniter.Context) {
        c.String(200, "Hello Godeniter Daemon!")
    })

    // 由守护进程管理器统一接管服务启动与生命周期指令
    _ = daemon.Run(app, ":8080", daemon.Config{
        Daemon:  false,            // 默认前台开发，通过 start 参数或改为 true 即可后台运行
        PIDFile: "./app.pid",
        LogFile: "./app.log",
    })
}
```

---

## 🐧 Linux 生产环境 Systemd 服务脚本推荐

虽然 Godeniter 原生支持 Daemon 模式，在成熟的企业级 Linux 服务器中，推荐将可执行文件托管给操作系统的 `systemd` 服务管理器：

```ini
# /etc/systemd/system/godeniter.service
[Unit]
Description=Godeniter Web Application
After=network.target

[Service]
Type=simple
User=www-data
WorkingDirectory=/var/www/godeniter-starter
ExecStart=/var/www/godeniter-starter/dist/app
Restart=always
RestartSec=5s
LimitNOFILE=65535

[Install]
WantedBy=multi-user.target
```

```bash
# 重载并开机自启
systemctl daemon-reload
systemctl enable --now godeniter
systemctl status godeniter
```

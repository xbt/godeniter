# 📴 离线环境与受限网络开发指南 (Offline & Air-gapped Guide)

**Godeniter** 采用 **100% 纯 Go 标准库（0 外部第三方依赖）** 设计。即使在**完全断网（内网机房、保密网络、无法访问 GitHub 的受限环境）**下，只要下载或拷贝了源码 Zip 包，也能获得极致流畅的开发与交付体验。

---

## 🎯 核心技术保障：为什么能 100% 纯离线运行？

很多现代 Web 框架（如 Gin、Fiber、Echo 等）或 ORM 框架在执行编译时，都需要依赖网络拉取几十甚至上百个第三方包。如果网络受限或断网，执行 `go run` 或 `go build` 就会报错中断。

而 **Godeniter**：
* 核心引擎完全基于 Go 1.20+ 原生标准库（`net/http`, `reflect`, `database/sql`, `html/template`, `crypto`, `embed` 等）实现。
* 根目录 `go.mod` 没有任何外部第三方依赖包声明。
* **无需执行 `go get`**，解压即可直接编译运行！

---

## 🚀 离线开发与使用的三种标准姿势

---

### 姿势一：配合 `godeniter-starter` 模板工程 (开箱即用，强烈推荐) ⭐

如果您需要快速启动一个全新的业务系统，建议同时下载 `godeniter` 框架源码包与 `godeniter-starter` 脚手架源码包，解压在**同一个父目录下**：

```text
my_workspace/
├── godeniter/           <-- 解压后的框架源码目录
└── godeniter-starter/   <-- 解压后的脚手架工程目录
```

#### 为什么无需联网？
在 `godeniter-starter/go.mod` 中已经内置了 Go 官方的本地路径替换指令：
```go
module godeniter-starter

go 1.20

require github.com/xbt/godeniter v0.0.0

// 优先直接读取同级目录下的本地源码，不发任何网络请求
replace github.com/xbt/godeniter => ../godeniter
```

#### 离线运行方式：
```bash
cd godeniter-starter

# 极速本地启动 (无需任何外网)
go run main.go
# 浏览器访问: http://127.0.0.1:8080 (带管理后台、Session 登录与 API 演示)
```

---

### 姿势二：在新建的任意空项目中本地引用 (`replace`)

如果您要在任意新目录（例如 `/data/my-project`）开始一个全新项目，也可以使用 Go 原生支持的 `replace` 特性指向本地解压的 `godeniter` 源码：

#### 1. 初始化新模块
```bash
mkdir my-project && cd my-project
go mod init my-project
```

#### 2. 添加本地替换指令
```bash
# 指向本地解压的 godeniter 路径 (支持相对路径或绝对路径)
go mod edit -replace github.com/xbt/godeniter=../path/to/godeniter
```

#### 3. 编写代码并直接运行
新建 `main.go`：
```go
package main

import (
    "github.com/xbt/godeniter"
)

func main() {
    app := godeniter.Classic()
    app.Get("/", func(c *godeniter.Context) {
        c.String(200, "📴 纯内网离线环境运行成功！")
    })
    app.Run(":8080")
}
```

执行启动：
```bash
go run main.go
```
Go 编译器将直接从本地读取 `godeniter` 源码并完成编译，全程不需要访问外网。

---

### 姿势三：直接交付单文件可执行程序 (`.exe`) —— 终极交付方案 📦

当您的项目开发完成后，如果最终客户机**没有网络**，甚至**没有安装任何 Go 开发环境**：

#### 1. 在开发机上一键编译单文件
```bash
# 在 Linux / macOS 或 Windows 开发机执行编译脚本
./build.sh
# 产物生成在: dist/app.exe (Windows 64位独立可执行程序)
```

#### 2. 客户机双击即用
* 将 `dist/app.exe` 直接通过 U 盘或内网文件传输拷贝给客户。
* 客户在 Windows 机器上**直接双击运行**。
* 程序已将前端 HTML/CSS 页面、API 接口和路由服务全部内嵌于单个二进制中，做到 **0 安装、0 网络依赖、双击即可启动**！

---

## 📊 离线场景支持矩阵对比

| 场景需求 | 外网连接 | Go 环境 | 第三方依赖拉取 | 离线支持体验 |
| :--- | :---: | :---: | :---: | :--- |
| **运行官方 Demo** | ❌ 不需要 | ✅ 需要 | ❌ 无任何第三方包 | 🟢 **秒级启动** (`go run ./examples/01_api_spa/main.go`) |
| **基于 Starter 开发** | ❌ 不需要 | ✅ 需要 | ❌ 无任何第三方包 | 🟢 **开箱即用** (`replace` 机制自动生效) |
| **全新项目本地引用** | ❌ 不需要 | ✅ 需要 | ❌ 无任何第三方包 | 🟢 **原生支持** (`go mod edit -replace ...`) |
| **客户机成品交付** | ❌ 不需要 | ❌ **不需要** | ❌ 无需环境配置 | 🟢 **双击单文件直接运行** (`.exe`) |

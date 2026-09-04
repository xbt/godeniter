# 动态配置管理开发手册 (Configuration & Database Setup)

在实际企业级项目和分发交付的软件中，**端口号、运行环境、数据库连接、密钥等配置不能硬编码在代码中**。
Godeniter 秉承 **100% 纯 Go 标准库（0 外部第三方依赖）** 原则，推荐使用 **`config.json`** 作为核心配置文件，配合“**三层动态装配机制**”，实现极速灵活的配置管理。

---

## 🎯 核心架构：三层动态覆盖体系

```text
1. 代码内置默认值 (Default Struct)   --> 保障即使没有配置文件也能安全启动
       ↓ (覆盖)
2. 外部配置文件 (config.json)        --> 允许客户机/运维使用文本编辑器随时修改端口与数据库
       ↓ (覆盖)
3. 系统环境变量 (OS Environment)     --> 适配 Docker 容器化、K8s 与云原生部署
```

---

## 📄 标准配置文件模板 (`config.json`)

在项目根目录下放置 `config.json`：

```json
{
  "app": {
    "name": "Godeniter 企业业务系统",
    "port": ":8080",
    "env": "development",
    "session_key": "godeniter-secure-session-salt-2026"
  },
  "database": {
    "driver": "sqlite",
    "dsn": "./data/app.db",
    "max_open_conns": 25,
    "max_idle_conns": 5,
    "conn_max_lifetime": 300
  },
  "upload": {
    "dir": "./uploads",
    "max_size_mb": 10,
    "allowed_exts": [
      ".jpg",
      ".png",
      ".jpeg",
      ".webp",
      ".pdf"
    ]
  }
}
```

---

## 💻 纯标准库配置解析器实现 (`config/config.go`)

无需引入任何第三方 YAML 或 Viper 库，仅用标准库 `encoding/json` 与 `os` 即可实现功能强大的配置管理器：

```go
package config

import (
    "encoding/json"
    "fmt"
    "os"
)

type Config struct {
    App      AppConfig      `json:"app"`
    Database DatabaseConfig `json:"database"`
    Upload   UploadConfig   `json:"upload"`
}

type AppConfig struct {
    Name       string `json:"name"`
    Port       string `json:"port"`
    Env        string `json:"env"`
    SessionKey string `json:"session_key"`
}

type DatabaseConfig struct {
    Driver          string `json:"driver"`
    DSN             string `json:"dsn"`
    MaxOpenConns    int    `json:"max_open_conns"`
    MaxIdleConns    int    `json:"max_idle_conns"`
    ConnMaxLifetime int    `json:"conn_max_lifetime"`
}

type UploadConfig struct {
    Dir         string   `json:"dir"`
    MaxSizeMB   int64    `json:"max_size_mb"`
    AllowedExts []string `json:"allowed_exts"`
}

// LoadConfig 智能加载配置
func LoadConfig(filePath ...string) *Config {
    cfg := DefaultConfig() // 1. 加载代码默认值

    path := "config.json"
    if len(filePath) > 0 && filePath[0] != "" {
        path = filePath[0]
    }

    // 2. 读取外部文件 (若不存在则自动生成模板)
    if data, err := os.ReadFile(path); err == nil {
        _ = json.Unmarshal(data, cfg)
    } else if os.IsNotExist(err) {
        if indentBytes, dumpErr := json.MarshalIndent(cfg, "", "  "); dumpErr == nil {
            _ = os.WriteFile(path, indentBytes, 0644)
            fmt.Printf(">> [CONFIG] 首次运行未检测到配置，已自动生成默认文件: %s\n", path)
        }
    }

    // 3. 环境变量最高优先级覆盖 (云原生/容器友好)
    if port := os.Getenv("PORT"); port != "" {
        cfg.App.Port = port
    }
    if dsn := os.Getenv("DATABASE_DSN"); dsn != "" {
        cfg.Database.DSN = dsn
    }

    return cfg
}
```

---

## 🗄️ 数据库连接动态初始化

在主入口 `main.go` 中，根据读取到的配置动态连接数据库，并注入到依赖容器中：

```go
package main

import (
    "github.com/xbt/godeniter"
    "github.com/xbt/godeniter/db"
    _ "modernc.org/sqlite" // 纯 Go SQLite 驱动 (或 _ "github.com/go-sql-driver/mysql")
    "my-project/config"
)

func main() {
    cfg := config.LoadConfig()

    app := godeniter.Classic()

    // 动态初始化数据库
    if cfg.Database.DSN != "" {
        database, err := db.Open(cfg.Database.Driver, cfg.Database.DSN)
        if err != nil {
            panic(fmt.Sprintf("数据库连接失败: %v", err))
        }
        database.SetMaxOpenConns(cfg.Database.MaxOpenConns)
        database.SetMaxIdleConns(cfg.Database.MaxIdleConns)

        // 注入全局依赖容器，Handler 任意使用！
        app.Map(database)
    }

    // 动态监听端口
    app.Run(cfg.App.Port)
}
```

---

## 📦 Windows 单文件交付时的使用场景

当打包为 `dist/app.exe` 交付给最终客户机时：

1. **客户双击 `app.exe` 首次启动**：
   * 程序检测到当前目录没有 `config.json`，直接以内置默认端口（如 `:8080`）启动；
   * 同时在 `app.exe` 同级目录下**自动生成了一份 `config.json`**。
2. **客户修改端口与数据库**：
   * 客户直接右键用记事本打开 `config.json`。
   * 将 `"port": ":8080"` 改为 `"port": ":9090"`。
   * 保存后重启 `app.exe`，服务立刻在新端口运行！
   * 真正做到了 **零环境依赖、单文件分发、动态外部配置**。

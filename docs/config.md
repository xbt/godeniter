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

## 🗄️ 数据库连接动态初始化 (SQLite 与 MySQL 示例)

在 `config.json` 中，可自由配置驱动与连接串：

### (1) MySQL 生产配置范例 (`config.json`)
```json
{
  "app": {
    "name": "Godeniter Demo",
    "port": ":8080",
    "env": "production"
  },
  "database": {
    "driver": "mysql",
    "dsn": "root:123456@tcp(127.0.0.1:3306)/godeniter_demo?charset=utf8mb4&parseTime=True&loc=Local",
    "max_open_conns": 50,
    "max_idle_conns": 10,
    "conn_max_lifetime": 3600
  }
}
```

### (2) SQLite 单文件配置范例 (`config.json`)
```json
{
  "database": {
    "driver": "sqlite",
    "dsn": "./data/app.db",
    "max_open_conns": 1,
    "max_idle_conns": 1,
    "conn_max_lifetime": 300
  }
}
```

### (3) 在主入口 `main.go` 中动态连接与注入容器
在业务代码中，只需根据配置载入连接，无需修改任何代码逻辑：

```go
package main

import (
    "fmt"
    "time"

    "github.com/xbt/godeniter"
    "github.com/xbt/godeniter/db"
    _ "github.com/go-sql-driver/mysql" // MySQL 驱动
    // _ "modernc.org/sqlite"          // SQLite 驱动 (按需启用)
    "my-project/config"
)

func main() {
    cfg := config.LoadConfig()
    app := godeniter.Classic()

    // 动态初始化数据库 (MySQL 或 SQLite 统一由 config.json 驱动)
    if cfg.Database.DSN != "" {
        database, err := db.Open(cfg.Database.Driver, cfg.Database.DSN)
        if err != nil {
            fmt.Printf(">> [WARN] 数据库连接失败: %v\n", err)
        } else {
            if cfg.Database.MaxOpenConns > 0 {
                database.SetMaxOpenConns(cfg.Database.MaxOpenConns)
            }
            if cfg.Database.MaxIdleConns > 0 {
                database.SetMaxIdleConns(cfg.Database.MaxIdleConns)
            }
            if cfg.Database.ConnMaxLifetime > 0 {
                database.SetConnMaxLifetime(time.Duration(cfg.Database.ConnMaxLifetime) * time.Second)
            }

            // 测试连通性
            if err := database.Ping(); err != nil {
                fmt.Printf(">> [WARN] 数据库 Ping 连通性测试失败: %v\n", err)
            } else {
                fmt.Printf(">> [DB] 数据库连接就绪 [%s -> %s]\n", cfg.Database.Driver, cfg.Database.DSN)
            }

            // 注入全局依赖容器，任意 Handler 可直接声明注入 *db.DB
            app.Map(database)
        }
    }

    // 动态监听端口
    _ = app.Run(cfg.App.Port)
}
```

---

## 💡 现场运维场景：客户机动态修改配置 (免重新编译)

交付可执行文件到生产服务器或客户机后，若需要修改端口或数据库连接，无需重新编译：

1. **首次启动自动就近生成**：
   * 程序检测到当前目录没有 `config.json` 时，自动在可执行文件旁生成一份标准格式的 `config.json`；
2. **修改端口与数据库**：
   * 用记事本或 Vim 打开 `config.json`；
   * 修改 `"port": ":9090"` 或修改 `"dsn": "user:pwd@tcp(192.168.1.100:3306)/dbname..."`；
   * 保存并重启程序，服务即刻以新端口与新数据库上线生效；
3. **云原生环境优先**：
   * 在 Docker / K8s 环境中，直接传入环境变量 `PORT=:8080`、`DATABASE_DSN="..."` 即可无感覆盖。

> ℹ️ **关于程序构建与二进制打包**：请参阅独立的 [《跨平台单文件打包与交付手册 (docs/build_and_deploy.md)》](./build_and_deploy.md)，开发阶段无需关注打包，专注于业务代码编写即可。

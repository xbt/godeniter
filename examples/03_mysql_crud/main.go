package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/xbt/godeniter"
	"github.com/xbt/godeniter/db"
	"github.com/xbt/godeniter/middleware"
	_ "github.com/go-sql-driver/mysql"
)

// Config 应用配置结构
type Config struct {
	App struct {
		Name string `json:"name"`
		Port string `json:"port"`
		Env  string `json:"env"`
	} `json:"app"`
	Database struct {
		Driver          string `json:"driver"`
		DSN             string `json:"dsn"`
		MaxOpenConns    int    `json:"max_open_conns"`
		MaxIdleConns    int    `json:"max_idle_conns"`
		ConnMaxLifetime int    `json:"conn_max_lifetime"`
	} `json:"database"`
}

// Article 文章数据实体 (映射 MySQL 中的 articles 表)
type Article struct {
	ID        int       `db:"id"         json:"id"`
	Title     string    `db:"title"      json:"title"`
	Content   string    `db:"content"    json:"content"`
	Author    string    `db:"author"     json:"author"`
	Views     int       `db:"views"      json:"views"`
	Status    int       `db:"status"     json:"status"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}

// CreateArticleDTO 新增文章请求入参及表单验证规则
type CreateArticleDTO struct {
	Title   string `json:"title"   binding:"required,min=2,max=100"`
	Content string `json:"content" binding:"required,min=5"`
	Author  string `json:"author"  binding:"required"`
}

// UpdateArticleDTO 更新文章请求入参
type UpdateArticleDTO struct {
	Title   string `json:"title"`
	Content string `json:"content"`
}

func loadConfig() *Config {
	cfg := &Config{}
	cfg.App.Port = ":8080"
	cfg.Database.Driver = "mysql"
	cfg.Database.DSN = "root:123456@tcp(127.0.0.1:3306)/godeniter_demo?charset=utf8mb4&parseTime=True&loc=Local"
	cfg.Database.MaxOpenConns = 50
	cfg.Database.MaxIdleConns = 10
	cfg.Database.ConnMaxLifetime = 3600

	if data, err := os.ReadFile("config.json"); err == nil {
		_ = json.Unmarshal(data, cfg)
	}
	if envDSN := os.Getenv("MYSQL_DSN"); envDSN != "" {
		cfg.Database.DSN = envDSN
	}
	if envPort := os.Getenv("PORT"); envPort != "" {
		cfg.App.Port = envPort
	}
	return cfg
}

func main() {
	cfg := loadConfig()

	// 1. 初始化 Godeniter 经典引擎 (带 Logger, Recovery 与优雅停机)
	app := godeniter.Classic()
	app.Use(middleware.CORS())

	// 2. 初始化 MySQL 数据库连接池
	var database *db.DB
	var dbErr error

	if cfg.Database.DSN != "" {
		database, dbErr = db.Open("mysql", cfg.Database.DSN)
		if dbErr == nil {
			database.SetMaxOpenConns(cfg.Database.MaxOpenConns)
			database.SetMaxIdleConns(cfg.Database.MaxIdleConns)
			database.SetConnMaxLifetime(time.Duration(cfg.Database.ConnMaxLifetime) * time.Second)

			// 测试 MySQL 连通性
			if pingErr := database.Ping(); pingErr == nil {
				fmt.Println(">> [DB] 成功连接至 MySQL 数据库:", cfg.Database.DSN)

				// 自动初始化数据表（若不存在则创建）
				initSQL := `
				CREATE TABLE IF NOT EXISTS articles (
					id INT AUTO_INCREMENT PRIMARY KEY,
					title VARCHAR(255) NOT NULL,
					content TEXT NOT NULL,
					author VARCHAR(100) NOT NULL DEFAULT '匿名',
					views INT NOT NULL DEFAULT 0,
					status TINYINT NOT NULL DEFAULT 1,
					created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
					updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
				) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;`
				_, _ = database.Exec(initSQL)

				// 检查若为空表则插入测试种子数据
				var count int64
				count, _ = database.Table("articles").Count()
				if count == 0 {
					_, _ = database.Table("articles").Insert(map[string]any{
						"title":      "欢迎使用 Godeniter + MySQL 实战演示",
						"content":    "这是一个通过纯 Go 标准库与 MySQL 驱动实现的完整 CRUD 演示工程，支持分页、模糊搜索与事务。",
						"author":     "admin",
						"views":      10,
						"status":     1,
						"created_at": time.Now(),
						"updated_at": time.Now(),
					})
				}
			} else {
				fmt.Printf(">> [WARN] MySQL Ping 失败: %v (请检查 config.json 中的 DSN 配置或启动本地 MySQL 服务)\n", pingErr)
				database = nil
			}
		} else {
			fmt.Printf(">> [WARN] 打开 MySQL 连接失败: %v\n", dbErr)
		}
	}

	// 3. 将数据库实例注入全局依赖注入 (DI) 容器
	if database != nil {
		app.Map(database)
	}

	// 4. 内嵌交互式测试页面 (方便在浏览器直接操作验证 MySQL)
	app.Get("/", func(c *godeniter.Context) {
		statusHtml := `<span style="color:#2ecc71;font-weight:bold;">已成功连接 MySQL 数据库</span>`
		if database == nil {
			statusHtml = `<span style="color:#e74c3c;font-weight:bold;">未连接 MySQL（请确保 MySQL 运行并配置正确 DSN）</span>`
		}

		html := fmt.Sprintf(`<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <title>Godeniter 2.0 - MySQL CRUD 实战演示</title>
    <style>
        body { font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif; max-width: 900px; margin: 40px auto; padding: 0 20px; color: #2c3e50; }
        .card { border: 1px solid #e2e8f0; border-radius: 8px; padding: 20px; margin-bottom: 20px; box-shadow: 0 2px 4px rgba(0,0,0,0.05); }
        .badge { display: inline-block; padding: 4px 10px; border-radius: 4px; font-size: 13px; background: #edf2f7; }
        button { background: #3498db; color: white; border: none; padding: 8px 16px; border-radius: 4px; cursor: pointer; font-size: 14px; margin-right: 8px; }
        button:hover { background: #2980b9; }
        pre { background: #1e293b; color: #f8fafc; padding: 15px; border-radius: 6px; overflow-x: auto; font-size: 13px; max-height: 400px; }
        input, textarea { width: 100%%; padding: 8px; border: 1px solid #cbd5e1; border-radius: 4px; box-sizing: border-box; margin-top: 6px; margin-bottom: 12px; }
    </style>
</head>
<body>
    <h1>🐬 Godeniter 2.0 - MySQL 实战 CRUD 控制台</h1>
    <div class="card">
        <p><strong>数据库状态：</strong> %s</p>
        <p><strong>当前 DSN：</strong> <code>%s</code></p>
        <p><strong>技术亮点：</strong> 100%%%% 纯 Go 标准库架构、ActiveRecord 链式查询构造器、连接池自动复用、依赖注入注入控制器。</p>
    </div>

    <div class="card">
        <h3>🚀 快捷操作接口测试</h3>
        <button onclick="fetchArticles()">1. 分页检索文章列表 (GET /api/v1/articles)</button>
        <button onclick="createArticle()">2. 新增一篇测试文章 (POST /api/v1/articles)</button>
        <button onclick="viewDetail()">3. 查看详情并自增阅读量 (GET /api/v1/articles/1)</button>
        <button onclick="runTransaction()">4. 执行批量归档事务 (POST /api/v1/articles/archive)</button>
        <pre id="output">// 点击上方按钮发起请求，此处展示实时 MySQL 响应结果...</pre>
    </div>

    <script>
        async function fetchArticles() {
            const res = await fetch('/api/v1/articles?page=1&page_size=5');
            const data = await res.json();
            document.getElementById('output').textContent = JSON.stringify(data, null, 2);
        }
        async function createArticle() {
            const res = await fetch('/api/v1/articles', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({
                    title: '新发表的实战心得 ' + new Date().toLocaleTimeString(),
                    content: '基于 Godeniter QueryBuilder 操作 MySQL，无需繁重 ORM，体验如同 CodeIgniter 般顺滑。',
                    author: 'xbt'
                })
            });
            const data = await res.json();
            document.getElementById('output').textContent = JSON.stringify(data, null, 2);
        }
        async function viewDetail() {
            const res = await fetch('/api/v1/articles/1');
            const data = await res.json();
            document.getElementById('output').textContent = JSON.stringify(data, null, 2);
        }
        async function runTransaction() {
            const res = await fetch('/api/v1/articles/archive?author=admin', { method: 'POST' });
            const data = await res.json();
            document.getElementById('output').textContent = JSON.stringify(data, null, 2);
        }
    </script>
</body>
</html>`, statusHtml, cfg.Database.DSN)

		c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(html))
	})

	// 5. RESTful API 路由分组 (/api/v1/articles)
	api := app.Group("/api/v1/articles")
	{
		// 1. 分页模糊检索 (GET /api/v1/articles?keyword=xxx&page=1&page_size=10)
		api.Get("", func(c *godeniter.Context) {
			if database == nil {
				c.Fail(50001, "MySQL 未连接，请在 config.json 中配置有效的 DSN 并确保数据库服务已启动")
				return
			}

			keyword := c.Query("keyword")
			page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
			pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

			qb := database.Table("articles").Where("status = ?", 1)
			if keyword != "" {
				qb.Like("title", "%"+keyword+"%")
			}

			var list []Article
			pager, err := qb.OrderBy("id DESC").Paginate(page, pageSize, &list)
			if err != nil {
				c.Fail(50002, "查询文章列表失败: "+err.Error())
				return
			}

			c.Success(godeniter.H{
				"list":  list,
				"pager": pager,
			})
		})

		// 2. 文章详情查询与阅读量自增 (GET /api/v1/articles/:id)
		api.Get("/:id", func(c *godeniter.Context) {
			if database == nil {
				c.Fail(50001, "MySQL 未连接")
				return
			}
			id := c.Param("id")

			var article Article
			err := database.Table("articles").Where("id = ?", id).First(&article)
			if err != nil {
				c.Fail(40401, "文章不存在或已被删除")
				return
			}

			// 阅读量原子自增
			_, _ = database.Table("articles").Where("id = ?", id).Increment("views", 1)
			article.Views++

			c.Success(article)
		})

		// 3. 新增文章 (POST /api/v1/articles)
		api.Post("", func(c *godeniter.Context) {
			if database == nil {
				c.Fail(50001, "MySQL 未连接")
				return
			}

			var form CreateArticleDTO
			if err := c.BindAndValidate(&form); err != nil {
				c.Fail(40001, "表单验证失败: "+err.Error())
				return
			}

			insertID, err := database.Table("articles").Insert(map[string]any{
				"title":      form.Title,
				"content":    form.Content,
				"author":     form.Author,
				"views":      0,
				"status":     1,
				"created_at": time.Now(),
				"updated_at": time.Now(),
			})
			if err != nil {
				c.Fail(50003, "保存文章失败: "+err.Error())
				return
			}

			c.Success(godeniter.H{
				"id":      insertID,
				"message": "文章创建成功",
			})
		})

		// 4. 更新文章 (PUT /api/v1/articles/:id)
		api.Put("/:id", func(c *godeniter.Context) {
			if database == nil {
				c.Fail(50001, "MySQL 未连接")
				return
			}
			id := c.Param("id")

			var form UpdateArticleDTO
			_ = c.BindJSON(&form)

			updateData := map[string]any{"updated_at": time.Now()}
			if form.Title != "" {
				updateData["title"] = form.Title
			}
			if form.Content != "" {
				updateData["content"] = form.Content
			}

			rows, err := database.Table("articles").Where("id = ?", id).Update(updateData)
			if err != nil || rows == 0 {
				c.Fail(50004, "更新失败或记录未发生变更")
				return
			}

			c.Success("文章更新成功")
		})

		// 5. 逻辑软删除 (DELETE /api/v1/articles/:id)
		api.Delete("/:id", func(c *godeniter.Context) {
			if database == nil {
				c.Fail(50001, "MySQL 未连接")
				return
			}
			id := c.Param("id")

			// 软删除: 更新 status 为 0
			rows, err := database.Table("articles").Where("id = ?", id).Update(map[string]any{"status": 0})
			if err != nil || rows == 0 {
				c.Fail(50005, "删除失败或文章不存在")
				return
			}

			c.Success("文章已成功删除(已软删除)")
		})

		// 6. 事务操作示例 (POST /api/v1/articles/archive?author=xxx)
		api.Post("/archive", func(c *godeniter.Context) {
			if database == nil {
				c.Fail(50001, "MySQL 未连接")
				return
			}
			author := c.Query("author")
			if author == "" {
				c.Fail(40002, "请指定需要归档的 author 参数")
				return
			}

			err := database.Transaction(func(tx *db.Tx) error {
				// 在同一个数据库事务中执行原子批量更新
				_, err := tx.Table("articles").Where("author = ?", author).Update(map[string]any{
					"status":     0,
					"updated_at": time.Now(),
				})
				if err != nil {
					return err // 返回 error 将自动触发 ROLLBACK
				}
				return nil // 返回 nil 将自动触发 COMMIT
			})

			if err != nil {
				c.Fail(50006, "事务执行失败: "+err.Error())
				return
			}

			c.Success(fmt.Sprintf("已成功将作者 [%s] 的所有文章安全归档(事务保护)", author))
		})
	}

	// 6. 启动 Web 服务
	_ = app.Run(cfg.App.Port)
}

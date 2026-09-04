# Godeniter 2.0 - MySQL CRUD 实战演示工程

本示例演示如何使用 **Godeniter 2.0 框架** 接入 **MySQL** 关系型数据库，实现企业级业务开发中高频使用的增删改查（CRUD）、分页检索、模糊搜索、数据模型映射与数据库事务。

---

## 🌟 核心特性展示

1. **MySQL 连接池标准配置**：包含最大连接数、最大空闲数、连接最长生命周期与连通性自检（`database.Ping()`）。
2. **自动初始化表结构**：首次连接若数据库中无 `articles` 表，自动执行建表与种子数据填充。
3. **数据实体与 Struct Tag 映射**：使用 `db:"id"` 与 `json:"id"` 双标签完成数据库到 JSON 响应的无缝映射。
4. **CodeIgniter 风格 ActiveRecord 链式查询**：
   * 模糊搜索：`qb.Like("title", "%keyword%")`
   * 条件过滤：`qb.Where("status = ?", 1)`
   * 自动分页：`qb.OrderBy("id DESC").Paginate(page, pageSize, &list)`
   * 原子自增：`qb.Increment("views", 1)`
5. **ACID 数据库事务**：`database.Transaction(...)`，支持自动提交与异常/Panic 自动安全回滚。
6. **可视化 Web 控制台**：访问首页 `http://127.0.0.1:8080` 提供交互式按钮，一键体验各接口。

---

## 🚀 极速启动运行

### 1. 准备 MySQL 数据库
确保本地或远程有正在运行的 MySQL 服务，并创建好空数据库：
```sql
CREATE DATABASE IF NOT EXISTS `godeniter_demo` DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
```
*(也可以直接执行当前目录下的 [`schema.sql`](./schema.sql) 导入结构)*

### 2. 配置数据库连接
打开当前目录下的 `config.json`，修改您的 MySQL 用户名、密码与地址：
```json
{
  "database": {
    "driver": "mysql",
    "dsn": "root:你的密码@tcp(127.0.0.1:3306)/godeniter_demo?charset=utf8mb4&parseTime=True&loc=Local"
  }
}
```
*(也可以通过环境变量注入：`export MYSQL_DSN="root:123456@tcp(127.0.0.1:3306)/godeniter_demo..."`)*

### 3. 安装 MySQL 驱动并启动
```bash
# 1. 引入官方 MySQL 驱动
go get github.com/go-sql-driver/mysql

# 2. 启动服务
go run main.go
```

在浏览器中打开：**`http://127.0.0.1:8080`** 即可进入交互式控制台！

---

## 📡 RESTful API 接口清单

| 方法 | 请求路径 | 功能说明 | 核心参数/Body |
| :--- | :--- | :--- | :--- |
| **GET** | `/api/v1/articles` | 分页与关键词检索 | `?keyword=Go&page=1&page_size=10` |
| **GET** | `/api/v1/articles/:id` | 获取详情并自增阅读量 | 路由参数 `:id` |
| **POST**| `/api/v1/articles` | 创建文章 (带验证) | `{"title":"...","content":"...","author":"..."}` |
| **PUT** | `/api/v1/articles/:id` | 更新文章标题或正文 | `{"title":"新标题","content":"新内容"}` |
| **DELETE**| `/api/v1/articles/:id` | 逻辑软删除文章 | 路由参数 `:id` (置 status 为 0) |
| **POST**| `/api/v1/articles/archive` | 事务批量归档某作者所有文章 | `?author=admin` |

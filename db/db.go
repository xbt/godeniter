package db

import (
	"database/sql"
	"fmt"
	"time"
)

// DB 封装了标准库 *sql.DB，提供连接管理、事务支持与链式查询构造入口。
type DB struct {
	*sql.DB
	driverName string
}

// Open 打开一个数据库连接池。
// driverName: "sqlite", "sqlite3", "mysql" 等（需在使用处通过匿名导入引入对应驱动）。
// dataSourceName: 数据库连接字符串（例如 SQLite 为 "./data.db"，MySQL 为 "user:password@tcp(127.0.0.1:3306)/dbname?charset=utf8mb4"）。
func Open(driverName string, dataSourceName string) (*DB, error) {
	stdDB, err := sql.Open(driverName, dataSourceName)
	if err != nil {
		return nil, fmt.Errorf("db: 连接数据库失败: %w", err)
	}

	// 设置默认连接池优化参数
	stdDB.SetMaxOpenConns(50)
	stdDB.SetMaxIdleConns(10)
	stdDB.SetConnMaxLifetime(time.Hour)

	// 测试连通性
	if err := stdDB.Ping(); err != nil {
		return nil, fmt.Errorf("db: 无法连通数据库 (ping error): %w", err)
	}

	return &DB{
		DB:         stdDB,
		driverName: driverName,
	}, nil
}

// Table 开始针对指定数据表构建链式查询。
// 示例：
//
//	db.Table("users").Where("status = ?", 1).Find(&users)
func (db *DB) Table(tableName string) *QueryBuilder {
	return newQueryBuilder(db.DB, tableName)
}

// Transaction 执行一个受事务保护的闭包函数。
// 如果闭包返回 error 或发生 panic，将自动回滚事务 (Rollback)；如果成功执行则自动提交 (Commit)。
// 示例：
//
//	err := db.Transaction(func(tx *db.Tx) error {
//	    _, err := tx.Table("accounts").Where("id = ?", 1).Update(map[string]any{"balance": 100})
//	    return err
//	})
func (db *DB) Transaction(fn func(tx *Tx) error) (err error) {
	stdTx, err := db.DB.Begin()
	if err != nil {
		return fmt.Errorf("db: 开启事务失败: %w", err)
	}

	tx := &Tx{Tx: stdTx}

	defer func() {
		if r := recover(); r != nil {
			_ = stdTx.Rollback()
			panic(r) // 重新抛出 panic
		} else if err != nil {
			_ = stdTx.Rollback()
		} else {
			err = stdTx.Commit()
		}
	}()

	err = fn(tx)
	return err
}

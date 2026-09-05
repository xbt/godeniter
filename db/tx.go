// Copyright (c) 2026 xbt. All rights reserved.
// Godeniter is licensed under the GNU General Public License v3.0 (GPL-3.0).
// See LICENSE file in the project root for full license information.

package db

import (
	"database/sql"
)

// Tx 封装了标准库 *sql.Tx，支持在事务内执行链式 QueryBuilder 操作。
type Tx struct {
	*sql.Tx
}

// Table 开始在事务上下文中针对指定数据表构建链式查询。
func (tx *Tx) Table(tableName string) *QueryBuilder {
	return newQueryBuilder(tx.Tx, tableName)
}

// Copyright (c) 2026 xbt. All rights reserved.
// Godeniter is licensed under the GNU General Public License v3.0 (GPL-3.0).
// See LICENSE file in the project root for full license information.

package models

import "time"

// Article 文章数据实体模型
type Article struct {
	ID         int       `json:"id" db:"id"`
	Title      string    `json:"title" db:"title"`
	Content    string    `json:"content" db:"content"`
	Author     string    `json:"author" db:"author"`
	AuthorMask string    `json:"author_mask"` // 脱敏后的作者名称
	CoverURL   string    `json:"cover_url" db:"cover_url"`
	Views      int       `json:"views" db:"views"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
}

// CreateArticleRequest 创建文章请求体 DTO (支持 binding 标签自动校验)
type CreateArticleRequest struct {
	Title    string `json:"title" binding:"required,min=2,max=50"`
	Content  string `json:"content" binding:"required,min=5"`
	Author   string `json:"author" binding:"required"`
	CoverURL string `json:"cover_url"`
}

// ArticleQueryRequest 文章分页与模糊搜索查询参数 DTO
type ArticleQueryRequest struct {
	Keyword  string `form:"keyword"`
	Page     int    `form:"page"`
	PageSize int    `form:"page_size"`
}

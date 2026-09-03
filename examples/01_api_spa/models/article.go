package models

import "time"

// Article 文章数据实体模型
type Article struct {
	ID        int       `json:"id" db:"id"`
	Title     string    `json:"title" db:"title"`
	Content   string    `json:"content" db:"content"`
	Author    string    `json:"author" db:"author"`
	Views     int       `json:"views" db:"views"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// CreateArticleRequest 创建文章请求体 DTO (支持 binding 标签自动校验)
type CreateArticleRequest struct {
	Title   string `json:"title" binding:"required,min=2,max=50"`
	Content string `json:"content" binding:"required,min=5"`
	Author  string `json:"author" binding:"required"`
}

// UpdateArticleRequest 更新文章请求体 DTO
type UpdateArticleRequest struct {
	Title   string `json:"title"`
	Content string `json:"content"`
}

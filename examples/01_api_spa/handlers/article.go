package handlers

import (
	"godeniter"
	"godeniter/examples/01_api_spa/models"
	"godeniter/router"
	"net/http"
	"strconv"
	"sync"
	"time"
)

// ArticleService 内存/数据库业务服务层接口 (通过 DI 注入到控制器中)
type ArticleService struct {
	mu       sync.RWMutex
	articles map[int]*models.Article
	nextID   int
}

// NewArticleService 创建并初始化文章服务（预置样例数据）
func NewArticleService() *ArticleService {
	s := &ArticleService{
		articles: make(map[int]*models.Article),
		nextID:   1,
	}
	s.Create("Godeniter 2.0 框架发布", "一款纯标准库、支持依赖注入与单文件打包的现代化 Web 框架。", "Ben")
	s.Create("前后端分离 API 开发最佳实践", "基于 RESTful 规范与统一 JSON 响应结构的设计指南。", "Antigravity")
	return s
}

// List 获取文章列表
func (s *ArticleService) List() []*models.Article {
	s.mu.RLock()
	defer s.mu.RUnlock()
	list := make([]*models.Article, 0, len(s.articles))
	for _, a := range s.articles {
		list = append(list, a)
	}
	return list
}

// GetByID 根据 ID 获取文章
func (s *ArticleService) GetByID(id int) (*models.Article, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	a, ok := s.articles[id]
	return a, ok
}

// Create 新增文章
func (s *ArticleService) Create(title, content, author string) *models.Article {
	s.mu.Lock()
	defer s.mu.Unlock()
	a := &models.Article{
		ID:        s.nextID,
		Title:     title,
		Content:   content,
		Author:    author,
		Views:     0,
		CreatedAt: time.Now(),
	}
	s.articles[s.nextID] = a
	s.nextID++
	return a
}

// Delete 删除文章
func (s *ArticleService) Delete(id int) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.articles[id]; ok {
		delete(s.articles, id)
		return true
	}
	return false
}

// ==============================================================================
// 控制器方法 (既支持 Context 方式，也支持 Martini DI 自动注入方式)
// ==============================================================================

// ListArticles 获取文章列表 (Context 方式)
func ListArticles(c *godeniter.Context, svc *ArticleService) {
	articles := svc.List()
	c.Success(godeniter.H{
		"total": len(articles),
		"items": articles,
	})
}

// GetArticleDetail 获取文章详情 (Martini 风格：直接声明参数 router.Params 与 *ArticleService)
func GetArticleDetail(params router.Params, svc *ArticleService) (int, godeniter.H) {
	idStr := params.Get("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return http.StatusBadRequest, godeniter.H{"code": 40001, "message": "非法的文章 ID"}
	}

	article, found := svc.GetByID(id)
	if !found {
		return http.StatusNotFound, godeniter.H{"code": 40401, "message": "文章不存在"}
	}

	return http.StatusOK, godeniter.H{
		"code":    0,
		"message": "ok",
		"data":    article,
	}
}

// CreateArticle 创建文章 (Context 方式)
func CreateArticle(c *godeniter.Context, svc *ArticleService) {
	var req models.CreateArticleRequest
	if err := c.BindJSON(&req); err != nil {
		c.Fail(40002, "请求参数 JSON 格式不正确: "+err.Error())
		return
	}

	if req.Title == "" || req.Content == "" {
		c.Fail(40003, "文章标题和内容不能为空")
		return
	}
	if req.Author == "" {
		req.Author = "匿名用户"
	}

	article := svc.Create(req.Title, req.Content, req.Author)
	c.Success(article)
}

// DeleteArticle 删除文章
func DeleteArticle(c *godeniter.Context, svc *ArticleService) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.Fail(40001, "非法的文章 ID")
		return
	}

	if !svc.Delete(id) {
		c.Fail(40401, "文章不存在或已被删除")
		return
	}

	c.Success(godeniter.H{"deleted_id": id})
}

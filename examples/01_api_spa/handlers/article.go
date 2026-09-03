package handlers

import (
	"fmt"
	"godeniter"
	"godeniter/examples/01_api_spa/models"
	"godeniter/router"
	"godeniter/utils/str"
	"math"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

// ArticleService 业务服务层
type ArticleService struct {
	mu       sync.RWMutex
	articles map[int]*models.Article
	nextID   int
}

// NewArticleService 创建并初始化文章服务
func NewArticleService() *ArticleService {
	s := &ArticleService{
		articles: make(map[int]*models.Article),
		nextID:   1,
	}
	s.Create("Godeniter 2.0 框架发布", "一款纯标准库、支持依赖注入、文件上传与单文件打包的现代化 Web 框架。", "ben@example.com", "")
	s.Create("前后端分离 API 开发最佳实践", "基于 RESTful 规范与统一 JSON 响应结构、分页查询的设计指南。", "13800138000", "")
	s.Create("CodeIgniter 哲学在 Go 中的重塑", "保留极速简单的开发直觉，提供零依赖的强大生产力工具库。", "antigravity@godeniter.dev", "")
	return s
}

// ListPaginate 分页与模糊搜索查询
func (s *ArticleService) ListPaginate(keyword string, page, pageSize int) ([]*models.Article, int) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	matched := make([]*models.Article, 0)
	kw := strings.ToLower(strings.TrimSpace(keyword))

	for _, a := range s.articles {
		if kw == "" || strings.Contains(strings.ToLower(a.Title), kw) || strings.Contains(strings.ToLower(a.Content), kw) {
			// 克隆对象并附加脱敏信息
			clone := *a
			if strings.Contains(clone.Author, "@") {
				clone.AuthorMask = str.MaskEmail(clone.Author)
			} else if len(clone.Author) == 11 {
				clone.AuthorMask = str.MaskPhone(clone.Author)
			} else {
				clone.AuthorMask = clone.Author
			}
			matched = append(matched, &clone)
		}
	}

	total := len(matched)
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 5
	}

	start := (page - 1) * pageSize
	if start >= total {
		return []*models.Article{}, total
	}
	end := start + pageSize
	if end > total {
		end = total
	}

	return matched[start:end], total
}

// GetByID 获取详情并自增阅读量 (Increment)
func (s *ArticleService) GetByID(id int) (*models.Article, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	a, ok := s.articles[id]
	if ok {
		a.Views++ // 模拟 Increment 阅读量
	}
	return a, ok
}

// Create 新增文章
func (s *ArticleService) Create(title, content, author, coverURL string) *models.Article {
	s.mu.Lock()
	defer s.mu.Unlock()
	a := &models.Article{
		ID:        s.nextID,
		Title:     title,
		Content:   content,
		Author:    author,
		CoverURL:  coverURL,
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
// 控制器方法
// ==============================================================================

// ListArticles 获取文章列表 (支持模糊搜索 keyword 与分页 page, page_size)
func ListArticles(c *godeniter.Context, svc *ArticleService) {
	var query models.ArticleQueryRequest
	_ = c.BindQuery(&query)

	if query.Page < 1 {
		query.Page = 1
	}
	if query.PageSize < 1 {
		query.PageSize = 5
	}

	items, total := svc.ListPaginate(query.Keyword, query.Page, query.PageSize)
	totalPages := int(math.Ceil(float64(total) / float64(query.PageSize)))
	if totalPages == 0 && total > 0 {
		totalPages = 1
	}

	c.Success(godeniter.H{
		"items": items,
		"pagination": godeniter.H{
			"total":       total,
			"page":        query.Page,
			"page_size":   query.PageSize,
			"total_pages": totalPages,
			"has_next":    query.Page < totalPages,
			"has_prev":    query.Page > 1,
		},
	})
}

// GetArticleDetail 获取文章详情 (自增阅读量)
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

// CreateArticle 创建文章
func CreateArticle(c *godeniter.Context, svc *ArticleService) {
	var req models.CreateArticleRequest
	if err := c.BindAndValidate(&req); err != nil {
		c.Fail(40002, "参数校验失败: "+err.Error())
		return
	}

	article := svc.Create(req.Title, req.Content, req.Author, req.CoverURL)
	c.Success(article)
}

// UploadFile 上传封面图片处理接口
func UploadFile(c *godeniter.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		c.Fail(40010, "请选择要上传的文件: "+err.Error())
		return
	}

	opts := godeniter.UploadOptions{
		SaveDir:     "./uploads/images",
		MaxBytes:    5 * 1024 * 1024, // 最大 5MB
		AllowedExts: []string{".jpg", ".jpeg", ".png", ".gif", ".webp"},
		AutoRename:  true,
	}

	savedPath, err := c.SaveUploadedFileWithOptions(file, opts)
	if err != nil {
		c.Fail(40011, "文件上传失败: "+err.Error())
		return
	}

	// 转换为对外的访问 URL 相对路径
	urlPath := "/" + filepath.ToSlash(savedPath)

	c.Success(godeniter.H{
		"filename":   file.Filename,
		"saved_path": savedPath,
		"url":        urlPath,
		"size":       file.Size,
	})
}

// DeleteArticle 删除文章
func DeleteArticle(c *godeniter.Context, svc *ArticleService) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.Fail(40001, fmt.Sprintf("非法文章ID: %s", idStr))
		return
	}

	if !svc.Delete(id) {
		c.Fail(40401, "文章不存在或已被删除")
		return
	}

	c.Success(godeniter.H{"deleted_id": id})
}

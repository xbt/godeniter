package controllers

import (
	"godeniter"
	"godeniter/session"
	"godeniter/utils/str"
	"net/http"
	"strings"
)

// Project 项目数据实体
type Project struct {
	ID        int
	Code      string // 随机工单编号 (str.UUID 截取)
	Name      string
	Summary   string // 智能截断摘要 (str.Truncate)
	Owner     string
	OwnerMask string // 脱敏后的联系人 (str.MaskPhone)
	Status    string
}

// HomeController 首页控制器
type HomeController struct{}

// Index 渲染首页列表 (支持关键词过滤与 Session 读取)
func (ctrl *HomeController) Index(c *godeniter.Context, sess session.Session) {
	var username, avatar string
	if sess != nil {
		username = sess.GetString("user_session")
		avatar = sess.GetString("user_avatar")
	}

	keyword := strings.TrimSpace(c.Query("keyword"))

	// 模拟数据库数据
	allProjects := []Project{
		{ID: 1, Code: "PRJ-" + str.Substr(str.MD5("1"), 0, 8), Name: "ERP 企业进销存系统 (标准版)", Summary: str.Truncate("包含商品管理、出入库审核、财务对账与报表统计等完整核心模块。", 20), Owner: "13800138000", OwnerMask: str.MaskPhone("13800138000"), Status: "运行中"},
		{ID: 2, Code: "PRJ-" + str.Substr(str.MD5("2"), 0, 8), Name: "智慧仓储 WMS 客户端 (扫码终端)", Summary: str.Truncate("基于嵌入式单文件交付，支持离线缓存与双击即用运行。", 20), Owner: "13911112222", OwnerMask: str.MaskPhone("13911112222"), Status: "运行中"},
		{ID: 3, Code: "PRJ-" + str.Substr(str.MD5("3"), 0, 8), Name: "自动化数据同步网关 (MySQL/SQLite)", Summary: str.Truncate("实现多源数据库的实时增量数据同步与健康状态监测。", 20), Owner: "13766668888", OwnerMask: str.MaskPhone("13766668888"), Status: "运行中"},
		{ID: 4, Code: "PRJ-" + str.Substr(str.MD5("4"), 0, 8), Name: "开放平台 Open API 认证网关", Summary: str.Truncate("支持 HMAC-SHA256 签名校验与速率限流控制。", 20), Owner: "13588889999", OwnerMask: str.MaskPhone("13588889999"), Status: "已就绪"},
	}

	filtered := make([]Project, 0)
	for _, p := range allProjects {
		if keyword == "" || strings.Contains(strings.ToLower(p.Name), strings.ToLower(keyword)) {
			filtered = append(filtered, p)
		}
	}

	c.HTML(http.StatusOK, "index.html", godeniter.H{
		"Title":       "Godeniter 管理控制台",
		"CurrentUser": username,
		"Avatar":      avatar,
		"Keyword":     keyword,
		"Projects":    filtered,
		"TotalCount":  len(filtered),
	})
}

// UploadAvatar 处理用户头像上传
func (ctrl *HomeController) UploadAvatar(c *godeniter.Context, sess session.Session) {
	file, err := c.FormFile("avatar")
	if err != nil {
		c.String(http.StatusBadRequest, "获取上传文件失败: %v", err)
		return
	}

	opts := godeniter.UploadOptions{
		SaveDir:     "./uploads/avatars",
		MaxBytes:    2 * 1024 * 1024, // 2MB
		AllowedExts: []string{".jpg", ".png", ".jpeg"},
		AutoRename:  true,
	}

	savedPath, err := c.SaveUploadedFileWithOptions(file, opts)
	if err != nil {
		c.String(http.StatusInternalServerError, "保存头像失败: %v", err)
		return
	}

	if sess != nil {
		sess.Set("user_avatar", "/"+savedPath)
	}

	c.Redirect(http.StatusFound, "/")
}

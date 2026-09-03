package controllers

import (
	"godeniter"
	"net/http"
)

// Project 项目数据实体
type Project struct {
	ID     int
	Name   string
	Owner  string
	Status string
}

// HomeController 首页控制器
type HomeController struct{}

// Index 渲染首页列表
func (ctrl *HomeController) Index(c *godeniter.Context) {
	// 获取 Cookie 中记录的当前登录用户
	username, _ := c.Cookie("user_session")

	// 模拟从数据库或业务层获取的数据
	projects := []Project{
		{ID: 1, Name: "ERP 企业进销存系统", Owner: "Ben", Status: "运行中"},
		{ID: 2, Name: "智慧仓储 WMS 客户端", Owner: "研发一部", Status: "运行中"},
		{ID: 3, Name: "自动化数据同步网关", Owner: "运维部", Status: "运行中"},
	}

	c.HTML(http.StatusOK, "index.html", godeniter.H{
		"Title":       "Godeniter 管理控制台",
		"CurrentUser": username,
		"Projects":    projects,
	})
}

package controllers

import (
	"godeniter"
	"net/http"
)

// UserController 用户认证控制器
type UserController struct{}

// LoginForm 显示登录页面 (GET /login)
func (ctrl *UserController) LoginForm(c *godeniter.Context) {
	c.HTML(http.StatusOK, "login.html", godeniter.H{
		"Title": "管理员登录",
	})
}

// LoginSubmit 处理表单 POST 提交登录 (POST /login)
func (ctrl *UserController) LoginSubmit(c *godeniter.Context) {
	username := c.PostForm("username")
	password := c.PostForm("password")

	// 简单验证（在实际开发中可配合 db.Table("users").Where(...) 验证密码哈希）
	if username == "admin" && password == "123456" {
		// 设置 Session Cookie (有效期 2 小时)
		c.SetCookie("user_session", username, 7200, "/", "", false, true)
		// 重定向回首页
		c.Redirect(http.StatusFound, "/")
		return
	}

	// 登录失败，返回登录页并显示错误提示
	c.HTML(http.StatusOK, "login.html", godeniter.H{
		"Title":    "管理员登录",
		"Username": username,
		"Error":    "账号或密码错误，请重新输入！",
	})
}

// Logout 注销退出 (GET /logout)
func (ctrl *UserController) Logout(c *godeniter.Context) {
	// 清除 Cookie
	c.SetCookie("user_session", "", -1, "/", "", false, true)
	c.Redirect(http.StatusFound, "/")
}

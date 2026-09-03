package models

// User 用户数据模型 (支持 db tag 映射)
type User struct {
	ID       int    `db:"id"`
	Username string `db:"username"`
	Password string `db:"password"`
	Role     string `db:"role"`
}

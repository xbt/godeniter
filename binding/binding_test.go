package binding

import (
	"bytes"
	"mime/multipart"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

type UserRegisterDTO struct {
	Username string `json:"username" form:"username" binding:"required,min=4,max=16"`
	Email    string `json:"email" form:"email" binding:"required,email"`
	Age      int    `json:"age" form:"age" binding:"required,min=18,max=100"`
	Phone    string `json:"phone" form:"phone" binding:"numeric"`
}

func TestValidator_Valid(t *testing.T) {
	dto := UserRegisterDTO{
		Username: "ben_dev",
		Email:    "ben@example.com",
		Age:      28,
		Phone:    "13800138000",
	}

	if err := Validate(&dto); err != nil {
		t.Fatalf("正常数据校验应当通过，但报错: %v", err)
	}
}

func TestValidator_Errors(t *testing.T) {
	// 测试必填与长度不足
	dto := UserRegisterDTO{
		Username: "abc",           // 长度小于 min=4
		Email:    "invalid-email", // 邮箱格式非法
		Age:      16,              // 小于 min=18
		Phone:    "138abc",        // 非纯数字
	}

	err := Validate(&dto)
	if err == nil {
		t.Fatalf("非法数据校验应当失败，但未报错")
	}

	ve, ok := err.(*ValidationError)
	if !ok {
		t.Fatalf("期望 ValidationError 类型，收到: %T", err)
	}

	if len(ve.Errors) != 4 {
		t.Errorf("期望 4 个字段错误，实际收到 %d 个: %v", len(ve.Errors), ve.Errors)
	}
}

func TestBindJSON(t *testing.T) {
	jsonBody := `{"username":"godeniter","email":"admin@godeniter.dev","age":25,"phone":"10086"}`
	req := httptest.NewRequest("POST", "/register", bytes.NewBufferString(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	var dto UserRegisterDTO
	if err := BindJSON(req, &dto); err != nil {
		t.Fatalf("BindJSON 失败: %v", err)
	}

	if dto.Username != "godeniter" || dto.Age != 25 {
		t.Errorf("BindJSON 字段解析错误: %+v", dto)
	}
}

func TestBindQuery(t *testing.T) {
	req := httptest.NewRequest("GET", "/search?username=tester&email=test@test.com&age=20&phone=123", nil)

	var dto UserRegisterDTO
	if err := BindQuery(req, &dto); err != nil {
		t.Fatalf("BindQuery 失败: %v", err)
	}

	if dto.Username != "tester" || dto.Email != "test@test.com" || dto.Age != 20 {
		t.Errorf("BindQuery 解析错误: %+v", dto)
	}
}

func TestBindForm(t *testing.T) {
	formData := url.Values{}
	formData.Set("username", "form_user")
	formData.Set("email", "form@user.com")
	formData.Set("age", "30")
	formData.Set("phone", "999999")

	req := httptest.NewRequest("POST", "/submit", strings.NewReader(formData.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	var dto UserRegisterDTO
	if err := BindForm(req, &dto); err != nil {
		t.Fatalf("BindForm 失败: %v", err)
	}

	if dto.Username != "form_user" || dto.Age != 30 {
		t.Errorf("BindForm 解析错误: %+v", dto)
	}
}

func TestBindMultipartForm(t *testing.T) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	_ = writer.WriteField("username", "multipart_user")
	_ = writer.WriteField("email", "mp@user.com")
	_ = writer.WriteField("age", "35")
	_ = writer.WriteField("phone", "1234567")

	part, err := writer.CreateFormFile("avatar", "avatar.png")
	if err != nil {
		t.Fatalf("CreateFormFile 失败: %v", err)
	}
	_, _ = part.Write([]byte("fake image data"))
	_ = writer.Close()

	req := httptest.NewRequest("POST", "/upload_submit", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	var dto UserRegisterDTO
	if err := Bind(req, &dto); err != nil {
		t.Fatalf("Bind multipart/form-data 失败: %v", err)
	}

	if dto.Username != "multipart_user" || dto.Email != "mp@user.com" || dto.Age != 35 || dto.Phone != "1234567" {
		t.Errorf("Bind multipart/form-data 字段解析错误: %+v", dto)
	}
}


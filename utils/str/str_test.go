package str

import (
	"strings"
	"testing"
)

func TestRandomGenerators(t *testing.T) {
	// 测试数字随机串
	num := RandomNumeric(6)
	if len(num) != 6 {
		t.Errorf("RandomNumeric(6) 长度错误: %s", num)
	}
	for _, r := range num {
		if r < '0' || r > '9' {
			t.Errorf("RandomNumeric 包含非数字字符: %c", r)
		}
	}

	// 测试 UUID
	uuid := UUID()
	if len(uuid) != 36 || strings.Count(uuid, "-") != 4 {
		t.Errorf("UUID 格式错误: %s", uuid)
	}
}

func TestNamingConversions(t *testing.T) {
	// SnakeCase
	if SnakeCase("UserProfile") != "user_profile" {
		t.Errorf("SnakeCase 转换错误: %s", SnakeCase("UserProfile"))
	}
	if SnakeCase("userID") != "user_id" {
		t.Errorf("SnakeCase 转换错误: %s", SnakeCase("userID"))
	}

	// CamelCase
	if CamelCase("user_profile", false) != "userProfile" {
		t.Errorf("CamelCase(false) 错误: %s", CamelCase("user_profile", false))
	}
	if CamelCase("user_profile", true) != "UserProfile" {
		t.Errorf("CamelCase(true) 错误: %s", CamelCase("user_profile", true))
	}

	// KebabCase
	if KebabCase("UserProfile") != "user-profile" {
		t.Errorf("KebabCase 错误: %s", KebabCase("UserProfile"))
	}
}

func TestTextHelpers(t *testing.T) {
	// Truncate
	text := "Godeniter 是一款轻量级 Go Web 框架"
	truncated := Truncate(text, 10, "...")
	if truncated != "Godeniter ..." {
		t.Errorf("Truncate 结果错误: %s", truncated)
	}

	// Substr (中文按字截取)
	sub := Substr("你好Godeniter世界", 2, 9)
	if sub != "Godeniter" {
		t.Errorf("Substr 中文截取错误: %s", sub)
	}
}

func TestMasking(t *testing.T) {
	if MaskPhone("13812345678") != "138****5678" {
		t.Errorf("MaskPhone 错误: %s", MaskPhone("13812345678"))
	}
	if MaskEmail("admin@godeniter.dev") != "a***n@godeniter.dev" {
		t.Errorf("MaskEmail 错误: %s", MaskEmail("admin@godeniter.dev"))
	}
	if MaskIDCard("110101199003072345") != "1101**********2345" {
		t.Errorf("MaskIDCard 错误: %s", MaskIDCard("110101199003072345"))
	}
}

func TestHashAndSecurity(t *testing.T) {
	// MD5("123456") = "e10adc3949ba59abbe56e057f20f883e"
	if MD5("123456") != "e10adc3949ba59abbe56e057f20f883e" {
		t.Errorf("MD5 错误: %s", MD5("123456"))
	}

	// XSS
	xssInput := "<script>alert('xss');</script>"
	filtered := XSSFilter(xssInput)
	if strings.Contains(filtered, "<script>") {
		t.Errorf("XSS 过滤失败: %s", filtered)
	}
}

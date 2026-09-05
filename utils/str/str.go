// Copyright (c) 2026 xbt. All rights reserved.
// Godeniter is licensed under the GNU General Public License v3.0 (GPL-3.0).
// See LICENSE file in the project root for full license information.

// Package str 提供了类似 PHP CodeIgniter (string_helper / text_helper / security_helper) 的丰富字符串与安全辅助函数库。
// 100% 基于纯 Go 标准库实现，零外部依赖。
package str

import (
	"crypto/hmac"
	"crypto/md5"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"html"
	"io"
	"strings"
	"unicode"
)

const (
	CharsetNumeric      = "0123456789"
	CharsetAlphaLower   = "abcdefghijklmnopqrstuvwxyz"
	CharsetAlphaUpper   = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	CharsetAlpha        = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	CharsetAlphaNumeric = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
)

// Random 根据指定长度和字符集生成高安全性随机字符串。
// 若未指定 charset，默认使用大小写字母+数字集 (CharsetAlphaNumeric)。
func Random(length int, charset ...string) string {
	if length <= 0 {
		return ""
	}
	chars := CharsetAlphaNumeric
	if len(charset) > 0 && charset[0] != "" {
		chars = charset[0]
	}

	bytes := make([]byte, length)
	if _, err := io.ReadFull(rand.Reader, bytes); err != nil {
		return ""
	}

	for i, b := range bytes {
		bytes[i] = chars[b%byte(len(chars))]
	}
	return string(bytes)
}

// RandomNumeric 生成指定长度的纯数字随机验证码 (例如 6 位数字验证码)。
func RandomNumeric(length int) string {
	return Random(length, CharsetNumeric)
}

// RandomAlpha 生成指定长度的纯字母随机字符串。
func RandomAlpha(length int) string {
	return Random(length, CharsetAlpha)
}

// UUID 生成符合 RFC 4122 标准的 Version 4 UUID 字符串 (如: "f47ac10b-58cc-4372-a567-0e02b2c3d479")。
func UUID() string {
	uuid := make([]byte, 16)
	if _, err := io.ReadFull(rand.Reader, uuid); err != nil {
		return ""
	}
	uuid[6] = (uuid[6] & 0x0f) | 0x40 // Version 4
	uuid[8] = (uuid[8] & 0x3f) | 0x80 // Variant RFC4122

	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x",
		uuid[0:4], uuid[4:6], uuid[6:8], uuid[8:10], uuid[10:16])
}

// SnakeCase 将驼峰命名或普通字符串转换为蛇形命名 (例如: "UserProfile" -> "user_profile", "APIResponse" -> "api_response")。
func SnakeCase(s string) string {
	var builder strings.Builder
	runes := []rune(s)
	length := len(runes)

	for i := 0; i < length; i++ {
		r := runes[i]
		if unicode.IsUpper(r) {
			if i > 0 && (unicode.IsLower(runes[i-1]) || (i+1 < length && unicode.IsLower(runes[i+1]))) {
				builder.WriteByte('_')
			}
			builder.WriteRune(unicode.ToLower(r))
		} else if r == '-' || r == ' ' {
			builder.WriteByte('_')
		} else {
			builder.WriteRune(r)
		}
	}
	return builder.String()
}

// CamelCase 将下划线或短横线命名转换为驼峰命名。
// upperFirst 为 true 表示大驼峰 (PascalCase，如 "user_name" -> "UserName")；false 表示小驼峰 (如 "user_name" -> "userName")。
func CamelCase(s string, upperFirst bool) string {
	s = strings.ReplaceAll(s, "-", "_")
	s = strings.ReplaceAll(s, " ", "_")
	parts := strings.Split(s, "_")

	var builder strings.Builder
	for i, part := range parts {
		if part == "" {
			continue
		}
		runes := []rune(part)
		if i == 0 && !upperFirst {
			runes[0] = unicode.ToLower(runes[0])
		} else {
			runes[0] = unicode.ToUpper(runes[0])
		}
		builder.WriteString(string(runes))
	}
	return builder.String()
}

// KebabCase 将字符串转换为短横线连字符命名 (例如: "user_profile" / "UserProfile" -> "user-profile")。
func KebabCase(s string) string {
	return strings.ReplaceAll(SnakeCase(s), "_", "-")
}

// Truncate 智能按字符数截断字符串，若超出长度则追加后缀 (默认后缀为 "...")。
func Truncate(s string, maxLen int, suffix ...string) string {
	runes := []rune(s)
	if len(runes) <= maxLen {
		return s
	}

	tail := "..."
	if len(suffix) > 0 {
		tail = suffix[0]
	}

	return string(runes[:maxLen]) + tail
}

// Substr 安全截取 UTF-8 字符串子串 (支持中文，按字符数量而非字节计算)。
func Substr(s string, start int, length int) string {
	runes := []rune(s)
	total := len(runes)

	if start < 0 {
		start = total + start
		if start < 0 {
			start = 0
		}
	}
	if start >= total {
		return ""
	}

	end := start + length
	if end > total || length < 0 {
		end = total
	}

	return string(runes[start:end])
}

// MaskPhone 手机号码脱敏 (例如: "13812345678" -> "138****5678")。
func MaskPhone(phone string) string {
	runes := []rune(phone)
	if len(runes) < 7 {
		return phone
	}
	if len(runes) == 11 {
		return string(runes[:3]) + "****" + string(runes[7:])
	}
	return string(runes[:3]) + "****" + string(runes[len(runes)-4:])
}

// MaskEmail 电子邮箱脱敏 (例如: "user@example.com" -> "u***r@example.com")。
func MaskEmail(email string) string {
	parts := strings.SplitN(email, "@", 2)
	if len(parts) != 2 {
		return email
	}

	name := []rune(parts[0])
	domain := parts[1]

	if len(name) <= 2 {
		return string(name[:1]) + "***@" + domain
	}
	return string(name[:1]) + "***" + string(name[len(name)-1:]) + "@" + domain
}

// MaskIDCard 身份证号码脱敏 (保留前 4 位和后 4 位，中间显示 "*")。
func MaskIDCard(idCard string) string {
	runes := []rune(idCard)
	if len(runes) < 8 {
		return idCard
	}
	maskLen := len(runes) - 8
	return string(runes[:4]) + strings.Repeat("*", maskLen) + string(runes[len(runes)-4:])
}

// MD5 计算字符串的 32 位小写 MD5 哈希摘要。
func MD5(s string) string {
	h := md5.New()
	h.Write([]byte(s))
	return hex.EncodeToString(h.Sum(nil))
}

// SHA256 计算字符串的 SHA256 十六进制哈希摘要。
func SHA256(s string) string {
	h := sha256.New()
	h.Write([]byte(s))
	return hex.EncodeToString(h.Sum(nil))
}

// HMACSHA256 使用指定密钥计算 HMAC-SHA256 签名。
func HMACSHA256(s, key string) string {
	h := hmac.New(sha256.New, []byte(key))
	h.Write([]byte(s))
	return hex.EncodeToString(h.Sum(nil))
}

// XSSFilter 进行基础的 HTML 安全实体转义，防止 XSS 跨站脚本攻击。
func XSSFilter(s string) string {
	return html.EscapeString(s)
}

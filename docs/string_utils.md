# 字符串与安全辅助函数库手册 (`utils/str`)

对标 **PHP CodeIgniter** 的 `string_helper`, `text_helper`, `security_helper`。

---

## 1. 常用函数列表

```go
import "godeniter/utils/str"

// 1. 随机字符生成
str.Random(16)            // 生成16位随机字符串
str.RandomNumeric(6)     // 生成6位数字验证码 (如 "839201")
str.RandomAlpha(8)       // 生成8位纯字母随机字符串
str.UUID()               // 生成标准 RFC4122 v4 UUID (如 "f47ac10b-58cc-4372-a567-0e02b2c3d479")

// 2. 命名规范转换
str.SnakeCase("UserProfile")        // "user_profile"
str.CamelCase("user_profile", true) // "UserProfile" (大驼峰/PascalCase)
str.CamelCase("user_profile", false)// "userProfile" (小驼峰)
str.KebabCase("UserProfile")        // "user-profile"

// 3. 文本处理与中文截断
str.Truncate("这是一篇非常长的文章正文...", 10, "...") // "这是一篇非常长..."
str.Substr("你好Godeniter世界", 2, 9)               // "Godeniter" (支持 UTF-8 中文按字截取)

// 4. 敏感隐私数据脱敏
str.MaskPhone("13812345678")          // "138****5678"
str.MaskEmail("admin@godeniter.dev")  // "a***n@godeniter.dev"
str.MaskIDCard("110101199003072345") // "1101**********2345"

// 5. 哈希算法与 XSS 安全实体过滤
str.MD5("123456")                    // 32位小写 MD5
str.SHA256("123456")                 // SHA-256 十六进制
str.HMACSHA256("message", "secret")  // HMAC-SHA256 签名
str.XSSFilter("<script>alert(1)</script>") // 实体转义
```

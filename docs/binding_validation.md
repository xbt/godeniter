# 参数绑定与轻量验证器手册 (Binding & Validation)

Godeniter 内置了基于 Struct Tag 的纯标准库数据验证引擎，支持 JSON、URL Query 与 POST 表单多源绑定。

---

## 1. 支持的校验规则列表

| 标签规则 | 示例 | 说明 |
| :--- | :--- | :--- |
| `required` | `binding:"required"` | 必填项，字符串不能为空，数字不能为 0，指针不能为 nil |
| `min=N` | `binding:"min=6"` | 字符串最小长度、数字最小值、切片最小元素数 |
| `max=N` | `binding:"max=20"`| 字符串最大长度、数字最大值、切片最大元素数 |
| `email` | `binding:"email"` | 验证是否为合法的电子邮箱格式 |
| `numeric` | `binding:"numeric"` | 验证字符串是否全由数字组成 |

---

## 2. 代码示例

```go
type CreateUserRequest struct {
    Username string `json:"username" form:"username" binding:"required,min=4,max=16"`
    Email    string `json:"email"    form:"email"    binding:"required,email"`
    Age      int    `json:"age"      form:"age"      binding:"required,min=18"`
}

app.Post("/users", func(c *godeniter.Context) {
    var req CreateUserRequest
    
    // 一行代码自动根据 Content-Type 解析并执行 struct tag 规则校验
    if err := c.BindAndValidate(&req); err != nil {
        c.Fail(40001, "表单验证不通过: " + err.Error())
        return
    }

    c.Success(godeniter.H{"username": req.Username})
})
```

# 文件上传与安全存储开发手册 (Upload Helper)

Godeniter 内置了类似 **PHP CodeIgniter Upload 类** 的上传处理机制，通过纯 Go 原生库实现，无第三方黑盒依赖。

---

## 1. 快速上手

```go
package main

import (
    "github.com/xbt/godeniter"
    "github.com/xbt/godeniter/utils/upload"
)

func main() {
    app := godeniter.Classic()

    // 映射静态资源目录，使上传文件可在浏览器中预览
    app.Static("/uploads", "./uploads")

    app.Post("/api/upload", func(c *godeniter.Context) {
        // 1. 获取上传文件
        file, err := c.FormFile("avatar")
        if err != nil {
            c.Fail(40001, "请选择文件: "+err.Error())
            return
        }

        // 2. 配置安全上传规则
        opts := upload.Options{
            SaveDir:     "./uploads/avatars",                   // 存储目录 (自动递归创建)
            MaxBytes:    2 * 1024 * 1024,                      // 最大 2MB
            AllowedExts: []string{".jpg", ".png", ".jpeg"},    // 扩展名白名单
            AutoRename:  true,                                 // 自动重命名为时间戳+随机字符串
        }

        // 3. 执行校验与保存
        savedPath, err := c.SaveUploadedFileWithOptions(file, opts)
        if err != nil {
            c.Fail(40002, "上传失败: "+err.Error())
            return
        }

        // 4. 返回前端可访问的相对 URL
        c.Success(godeniter.H{
            "filename": file.Filename,
            "url":      "/" + savedPath,
        })
    })

    app.Run(":8080")
}
```

---

## 2. 核心 API 方法速查

| 方法签名 | 说明 |
| :--- | :--- |
| `c.FormFile(name string) (*multipart.FileHeader, error)` | 获取单文件上传对象 |
| `c.FormFiles(name string) ([]*multipart.FileHeader, error)` | 获取同名字段的多个上传文件列表 |
| `c.SaveUploadedFile(file, dstPath)` | 快捷保存文件到目标绝对/相对路径 |
| `c.SaveUploadedFileWithOptions(file, opts)` | 结合大小/扩展名规则校验后安全存储 |

// Copyright (c) 2026 xbt. All rights reserved.
// Godeniter is licensed under the GNU General Public License v3.0 (GPL-3.0).
// See LICENSE file in the project root for full license information.

// Package upload 提供了类似 PHP CodeIgniter Upload 类的文件上传与安全存储处理。
// 100% 纯 Go 标准库实现，支持文件大小、扩展名白名单校验与时间戳随机重命名。
package upload

import (
	"fmt"
	"github.com/xbt/godeniter/utils/str"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Options 定义文件上传处理与安全校验选项（参考 CodeIgniter Upload Config）。
type Options struct {
	SaveDir     string   // 保存目标目录 (如 "./uploads/images")，若不存在将自动递归创建
	MaxBytes    int64    // 最大允许文件大小 (字节)，如 5*1024*1024 (5MB)。0 表示不限制
	AllowedExts []string // 允许的文件扩展名白名单 (如 []string{".jpg", ".png", ".gif"})，忽略大小写
	AutoRename  bool     // 是否自动重命名为唯一安全文件名 (格式: YYYYMMDD_HHMMSS_随机串.ext)
}

// DefaultOptions 默认上传配置选项 (最大 10MB，自动重命名)。
var DefaultOptions = Options{
	SaveDir:    "./uploads",
	MaxBytes:   10 * 1024 * 1024,
	AutoRename: true,
}

// SaveUploadedFile 将上传的内存/临时文件保存到指定的目标文件绝对或相对路径。
func SaveUploadedFile(file *multipart.FileHeader, dst string) error {
	src, err := file.Open()
	if err != nil {
		return fmt.Errorf("upload: 打开上传源文件失败: %w", err)
	}
	defer src.Close()

	// 确保目标父级目录存在
	dir := filepath.Dir(dst)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("upload: 创建存储目录失败: %w", err)
	}

	out, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("upload: 创建目标文件失败: %w", err)
	}
	defer out.Close()

	_, err = io.Copy(out, src)
	return err
}

// SaveUploadedFileWithOptions 按照 Options 安全规则进行大小、扩展名校验并完成存储。
// 返回最终保存的文件相对路径。
func SaveUploadedFileWithOptions(file *multipart.FileHeader, opts Options) (string, error) {
	if file == nil {
		return "", fmt.Errorf("upload: 文件对象不能为空")
	}

	// 1. 文件大小校验
	if opts.MaxBytes > 0 && file.Size > opts.MaxBytes {
		return "", fmt.Errorf("upload: 文件大小 (%d 字节) 超出最大限制 (%d 字节)", file.Size, opts.MaxBytes)
	}

	// 2. 扩展名白名单校验
	ext := strings.ToLower(filepath.Ext(file.Filename))
	if len(opts.AllowedExts) > 0 {
		allowed := false
		for _, allowExt := range opts.AllowedExts {
			if strings.ToLower(allowExt) == ext || strings.ToLower("."+allowExt) == ext {
				allowed = true
				break
			}
		}
		if !allowed {
			return "", fmt.Errorf("upload: 不允许上传 [%s] 格式的文件，仅支持: %v", ext, opts.AllowedExts)
		}
	}

	// 3. 计算目标存储路径与文件名
	filename := file.Filename
	if opts.AutoRename {
		filename = fmt.Sprintf("%s_%s%s",
			time.Now().Format("20060102_150405"),
			str.Random(8, str.CharsetAlphaNumeric),
			ext,
		)
	}

	saveDir := opts.SaveDir
	if saveDir == "" {
		saveDir = "./uploads"
	}

	finalPath := filepath.Join(saveDir, filename)
	if err := SaveUploadedFile(file, finalPath); err != nil {
		return "", err
	}

	return filepath.ToSlash(finalPath), nil
}

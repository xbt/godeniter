// Copyright (c) 2026 xbt. All rights reserved.
// Godeniter is licensed under the GNU General Public License v3.0 (GPL-3.0).
// See LICENSE file in the project root for full license information.

// Package storage 提供纯 Go 标准库实现的多存储驱动抽象 (0 外部依赖，0 臃肿 SDK)。
// 支持本地文件系统 (Local)、HTTP WebDAV、以及原生 AWS SigV4 签名的 S3 / Cloudflare R2 / 阿里云 OSS / MinIO 对象存储。
package storage

import (
	"io"
)

// Driver 统一多存储驱动接口
type Driver interface {
	// Name 返回驱动名称 (local / webdav / s3 等)
	Name() string

	// Save 保存文件流并返回可访问的文件 URL 或相对路径
	Save(filename string, data io.Reader, size int64, contentType string) (fileURL string, err error)

	// Delete 删除已存储的文件
	Delete(filename string) error
}

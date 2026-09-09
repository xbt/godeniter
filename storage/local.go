// Copyright (c) 2026 xbt. All rights reserved.
// Godeniter is licensed under the GNU General Public License v3.0 (GPL-3.0).
// See LICENSE file in the project root for full license information.

package storage

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// LocalStorageDriver 本地文件系统存储驱动
type LocalStorageDriver struct {
	baseDir   string // 本地物理保存目录 (如 ./uploads)
	urlPrefix string // 暴露给 HTTP 访问的前缀 (如 /uploads)
}

// NewLocalStorageDriver 创建本地存储驱动实例
func NewLocalStorageDriver(baseDir string, urlPrefix ...string) *LocalStorageDriver {
	prefix := "/uploads"
	if len(urlPrefix) > 0 && urlPrefix[0] != "" {
		prefix = urlPrefix[0]
	}
	prefix = "/" + strings.Trim(prefix, "/")

	return &LocalStorageDriver{
		baseDir:   baseDir,
		urlPrefix: prefix,
	}
}

func (d *LocalStorageDriver) Name() string {
	return "local"
}

func (d *LocalStorageDriver) Save(filename string, data io.Reader, size int64, contentType string) (string, error) {
	cleanName := filepath.Base(filename)
	if err := os.MkdirAll(d.baseDir, 0755); err != nil {
		return "", fmt.Errorf("godeniter/storage: 创建本地目录失败: %w", err)
	}

	targetPath := filepath.Join(d.baseDir, cleanName)
	out, err := os.Create(targetPath)
	if err != nil {
		return "", fmt.Errorf("godeniter/storage: 创建文件失败: %w", err)
	}
	defer out.Close()

	if _, err := io.Copy(out, data); err != nil {
		return "", fmt.Errorf("godeniter/storage: 写入文件失败: %w", err)
	}

	return d.urlPrefix + "/" + cleanName, nil
}

func (d *LocalStorageDriver) Delete(filename string) error {
	cleanName := filepath.Base(filename)
	targetPath := filepath.Join(d.baseDir, cleanName)
	return os.Remove(targetPath)
}

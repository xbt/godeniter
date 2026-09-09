// Copyright (c) 2026 xbt. All rights reserved.
// Godeniter is licensed under the GNU General Public License v3.0 (GPL-3.0).
// See LICENSE file in the project root for full license information.

package storage

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strings"
	"time"
)

// WebDAVStorageDriver 基于 HTTP WebDAV 协议驱动 (支持 Nextcloud, 坚果云, InfiniCLOUD 等)
type WebDAVStorageDriver struct {
	serverURL string
	username  string
	password  string
	client    *http.Client
}

// NewWebDAVStorageDriver 创建 WebDAV 存储驱动实例
func NewWebDAVStorageDriver(serverURL, username, password string, customClient ...*http.Client) *WebDAVStorageDriver {
	cli := &http.Client{Timeout: 30 * time.Second}
	if len(customClient) > 0 && customClient[0] != nil {
		cli = customClient[0]
	}

	return &WebDAVStorageDriver{
		serverURL: strings.TrimRight(serverURL, "/"),
		username:  username,
		password:  password,
		client:    cli,
	}
}

func (d *WebDAVStorageDriver) Name() string {
	return "webdav"
}

func (d *WebDAVStorageDriver) Save(filename string, data io.Reader, size int64, contentType string) (string, error) {
	cleanName := filepath.Base(filename)
	targetURL := fmt.Sprintf("%s/%s", d.serverURL, cleanName)

	buf := new(bytes.Buffer)
	if _, err := io.Copy(buf, data); err != nil {
		return "", err
	}

	req, err := http.NewRequest(http.MethodPut, targetURL, buf)
	if err != nil {
		return "", err
	}

	if d.username != "" {
		req.SetBasicAuth(d.username, d.password)
	}
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}

	resp, err := d.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("godeniter/storage: webdav upload failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusNoContent {
		return "", fmt.Errorf("godeniter/storage: webdav server returned status %d", resp.StatusCode)
	}

	return targetURL, nil
}

func (d *WebDAVStorageDriver) Delete(filename string) error {
	cleanName := filepath.Base(filename)
	targetURL := fmt.Sprintf("%s/%s", d.serverURL, cleanName)

	req, err := http.NewRequest(http.MethodDelete, targetURL, nil)
	if err != nil {
		return err
	}
	if d.username != "" {
		req.SetBasicAuth(d.username, d.password)
	}

	resp, err := d.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("godeniter/storage: webdav delete failed with status %d", resp.StatusCode)
	}
	return nil
}

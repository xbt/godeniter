// Copyright (c) 2026 xbt. All rights reserved.
// Godeniter is licensed under the GNU General Public License v3.0 (GPL-3.0).
// See LICENSE file in the project root for full license information.

package upload

import (
	"bytes"
	"mime/multipart"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestUpload(t *testing.T) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile("avatar", "test_avatar.png")
	if err != nil {
		t.Fatalf("CreateFormFile 失败: %v", err)
	}
	part.Write([]byte("fake image content bytes"))
	writer.Close()

	req := httptest.NewRequest("POST", "/upload", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	file, fh, err := req.FormFile("avatar")
	if err != nil {
		t.Fatalf("req.FormFile 失败: %v", err)
	}
	file.Close()

	opts := Options{
		SaveDir:     "./dist/test_upload_util",
		MaxBytes:    1024 * 1024,
		AllowedExts: []string{".png", ".jpg"},
		AutoRename:  true,
	}

	saved, err := SaveUploadedFileWithOptions(fh, opts)
	if err != nil {
		t.Fatalf("SaveUploadedFileWithOptions 失败: %v", err)
	}

	if _, err := os.Stat(saved); os.IsNotExist(err) {
		t.Fatalf("保存后的文件不存在: %s", saved)
	}

	_ = os.RemoveAll(filepath.Dir(saved))
}

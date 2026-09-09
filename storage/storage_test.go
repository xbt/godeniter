// Copyright (c) 2026 xbt. All rights reserved.
// Godeniter is licensed under the GNU General Public License v3.0 (GPL-3.0).
// See LICENSE file in the project root for full license information.

package storage

import (
	"bytes"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// inMemoryRoundTripper 纯内存 Mock HTTP Transport，避免依赖系统 Socket/网络权限
type inMemoryRoundTripper struct {
	handler func(req *http.Request) (*http.Response, error)
}

func (m *inMemoryRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	return m.handler(req)
}

func TestLocalStorageDriver(t *testing.T) {
	tmpDir := t.TempDir()
	driver := NewLocalStorageDriver(tmpDir, "/custom_uploads")

	if driver.Name() != "local" {
		t.Fatalf("expected driver name 'local', got %s", driver.Name())
	}

	testData := []byte("hello godeniter storage")
	url, err := driver.Save("test.txt", bytes.NewReader(testData), int64(len(testData)), "text/plain")
	if err != nil {
		t.Fatalf("save failed: %v", err)
	}

	if url != "/custom_uploads/test.txt" {
		t.Fatalf("expected url '/custom_uploads/test.txt', got %s", url)
	}

	savedFile := filepath.Join(tmpDir, "test.txt")
	content, err := os.ReadFile(savedFile)
	if err != nil || string(content) != "hello godeniter storage" {
		t.Fatalf("file content mismatch: %v", err)
	}

	if err := driver.Delete("test.txt"); err != nil {
		t.Fatalf("delete failed: %v", err)
	}

	if _, err := os.Stat(savedFile); !os.IsNotExist(err) {
		t.Fatalf("file should be deleted")
	}
}

func TestWebDAVStorageDriver(t *testing.T) {
	var receivedMethod string
	var receivedAuth string
	var receivedBody string

	mockClient := &http.Client{
		Transport: &inMemoryRoundTripper{
			handler: func(req *http.Request) (*http.Response, error) {
				receivedMethod = req.Method
				receivedAuth = req.Header.Get("Authorization")
				buf := new(bytes.Buffer)
				_, _ = buf.ReadFrom(req.Body)
				receivedBody = buf.String()

				return &http.Response{
					StatusCode: http.StatusCreated,
					Body:       io.NopCloser(strings.NewReader("")),
					Header:     make(http.Header),
				}, nil
			},
		},
	}

	driver := NewWebDAVStorageDriver("https://webdav.example.com/remote.php/dav/files", "myuser", "mypass", mockClient)
	if driver.Name() != "webdav" {
		t.Fatalf("expected driver name 'webdav', got %s", driver.Name())
	}

	testData := "webdav-file-content"
	url, err := driver.Save("doc.txt", strings.NewReader(testData), int64(len(testData)), "text/plain")
	if err != nil {
		t.Fatalf("save to webdav failed: %v", err)
	}

	if receivedMethod != http.MethodPut {
		t.Errorf("expected PUT, got %s", receivedMethod)
	}
	if !strings.HasPrefix(receivedAuth, "Basic ") {
		t.Errorf("expected basic auth, got %s", receivedAuth)
	}
	if receivedBody != testData {
		t.Errorf("body mismatch: %s vs %s", receivedBody, testData)
	}
	if !strings.HasSuffix(url, "/doc.txt") {
		t.Errorf("url mismatch: %s", url)
	}
}

func TestS3StorageDriver_SigV4(t *testing.T) {
	var authHeader string
	var amzDate string
	var contentSha string

	mockClient := &http.Client{
		Transport: &inMemoryRoundTripper{
			handler: func(req *http.Request) (*http.Response, error) {
				authHeader = req.Header.Get("Authorization")
				amzDate = req.Header.Get("x-amz-date")
				contentSha = req.Header.Get("x-amz-content-sha256")

				return &http.Response{
					StatusCode: http.StatusOK,
					Body:       io.NopCloser(strings.NewReader("")),
					Header:     make(http.Header),
				}, nil
			},
		},
	}

	driver := NewS3StorageDriver(S3Config{
		Endpoint:  "s3.us-east-1.amazonaws.com",
		Region:    "us-east-1",
		Bucket:    "mybucket",
		AccessKey: "AKIAIOSFODNN7EXAMPLE",
		SecretKey: "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
		Domain:    "https://cdn.example.com",
		Client:    mockClient,
	})

	if driver.Name() != "s3" {
		t.Fatalf("expected driver name 's3', got %s", driver.Name())
	}

	testData := "s3 payload test"
	url, err := driver.Save("image.png", strings.NewReader(testData), int64(len(testData)), "image/png")
	if err != nil {
		t.Fatalf("save to s3 failed: %v", err)
	}

	if url != "https://cdn.example.com/image.png" {
		t.Fatalf("expected CDN url 'https://cdn.example.com/image.png', got %s", url)
	}
	if !strings.HasPrefix(authHeader, "AWS4-HMAC-SHA256 Credential=AKIAIOSFODNN7EXAMPLE/2026") && !strings.HasPrefix(authHeader, "AWS4-HMAC-SHA256 Credential=AKIAIOSFODNN7EXAMPLE/") {
		t.Errorf("invalid SigV4 auth header: %s", authHeader)
	}
	if amzDate == "" || contentSha == "" {
		t.Errorf("missing SigV4 headers: amzDate=%s, contentSha=%s", amzDate, contentSha)
	}
}

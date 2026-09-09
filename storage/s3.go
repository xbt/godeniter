// Copyright (c) 2026 xbt. All rights reserved.
// Godeniter is licensed under the GNU General Public License v3.0 (GPL-3.0).
// See LICENSE file in the project root for full license information.

package storage

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strings"
	"time"
)

// S3Config S3 兼容对象存储配置
type S3Config struct {
	Endpoint  string // 接口地址 (如 s3.us-east-1.amazonaws.com, xxx.r2.cloudflarestorage.com)
	Region    string // 区域 (默认 "auto")
	Bucket    string // 存储桶名称
	AccessKey string // 访问凭据 AccessKeyId
	SecretKey string // 访问密钥 SecretAccessKey
	Domain    string // 可选 CDN / 自定义访问域名 (如 https://cdn.example.com)
	Client    *http.Client
}

// S3StorageDriver 纯 Go 标准库实现的轻量 S3 驱动 (纯手写 AWS SigV4 规范签名算法，0 外部 SDK)
type S3StorageDriver struct {
	cfg    S3Config
	client *http.Client
}

// NewS3StorageDriver 创建 S3 存储驱动实例
func NewS3StorageDriver(cfg S3Config) *S3StorageDriver {
	if cfg.Region == "" {
		cfg.Region = "auto"
	}
	// 去除 endpoint 的协议前缀
	ep := strings.TrimPrefix(cfg.Endpoint, "https://")
	ep = strings.TrimPrefix(ep, "http://")
	cfg.Endpoint = strings.TrimRight(ep, "/")
	cfg.Domain = strings.TrimRight(cfg.Domain, "/")

	cli := cfg.Client
	if cli == nil {
		cli = &http.Client{Timeout: 30 * time.Second}
	}

	return &S3StorageDriver{
		cfg:    cfg,
		client: cli,
	}
}

func (d *S3StorageDriver) Name() string {
	return "s3"
}

func (d *S3StorageDriver) Save(filename string, data io.Reader, size int64, contentType string) (string, error) {
	cleanName := filepath.Base(filename)
	bodyBytes, err := io.ReadAll(data)
	if err != nil {
		return "", fmt.Errorf("godeniter/storage: 读取上传数据失败: %w", err)
	}

	if contentType == "" {
		contentType = "application/octet-stream"
	}

	reqURL := fmt.Sprintf("https://%s/%s/%s", d.cfg.Endpoint, d.cfg.Bucket, cleanName)
	req, err := http.NewRequest(http.MethodPut, reqURL, bytes.NewReader(bodyBytes))
	if err != nil {
		return "", err
	}

	now := time.Now().UTC()
	amzDate := now.Format("20060102T150405Z")
	dateStamp := now.Format("20060102")

	// Payload 哈希计算
	hPayload := sha256.Sum256(bodyBytes)
	payloadHash := hex.EncodeToString(hPayload[:])

	req.Header.Set("Host", d.cfg.Endpoint)
	req.Header.Set("x-amz-date", amzDate)
	req.Header.Set("x-amz-content-sha256", payloadHash)
	req.Header.Set("Content-Type", contentType)

	// 计算规范化请求 (CanonicalRequest)
	canonicalURI := fmt.Sprintf("/%s/%s", d.cfg.Bucket, cleanName)
	canonicalHeaders := fmt.Sprintf("content-type:%s\nhost:%s\nx-amz-content-sha256:%s\nx-amz-date:%s\n",
		contentType, d.cfg.Endpoint, payloadHash, amzDate)
	signedHeaders := "content-type;host;x-amz-content-sha256;x-amz-date"

	canonicalRequest := fmt.Sprintf("%s\n%s\n%s\n%s\n%s\n%s",
		http.MethodPut,
		canonicalURI,
		"", // query string
		canonicalHeaders,
		signedHeaders,
		payloadHash,
	)

	// 计算 StringToSign
	hCanonical := sha256.Sum256([]byte(canonicalRequest))
	credentialScope := fmt.Sprintf("%s/%s/s3/aws4_request", dateStamp, d.cfg.Region)
	stringToSign := fmt.Sprintf("AWS4-HMAC-SHA256\n%s\n%s\n%s",
		amzDate,
		credentialScope,
		hex.EncodeToString(hCanonical[:]),
	)

	// 计算 Signature
	signingKey := getSignatureKey(d.cfg.SecretKey, dateStamp, d.cfg.Region, "s3")
	signature := hmacSHA256(signingKey, []byte(stringToSign))

	// 注入 Authorization 标头
	authHeader := fmt.Sprintf("AWS4-HMAC-SHA256 Credential=%s/%s, SignedHeaders=%s, Signature=%s",
		d.cfg.AccessKey, credentialScope, signedHeaders, hex.EncodeToString(signature))
	req.Header.Set("Authorization", authHeader)

	resp, err := d.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("godeniter/storage: S3 上传请求失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusNoContent {
		respBody, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("godeniter/storage: S3 响应状态异常 (%d): %s", resp.StatusCode, string(respBody))
	}

	if d.cfg.Domain != "" {
		return fmt.Sprintf("%s/%s", d.cfg.Domain, cleanName), nil
	}
	return reqURL, nil
}

func (d *S3StorageDriver) Delete(filename string) error {
	cleanName := filepath.Base(filename)
	reqURL := fmt.Sprintf("https://%s/%s/%s", d.cfg.Endpoint, d.cfg.Bucket, cleanName)
	req, err := http.NewRequest(http.MethodDelete, reqURL, nil)
	if err != nil {
		return err
	}

	now := time.Now().UTC()
	amzDate := now.Format("20060102T150405Z")
	dateStamp := now.Format("20060102")

	hEmpty := sha256.Sum256([]byte(""))
	payloadHash := hex.EncodeToString(hEmpty[:])

	req.Header.Set("Host", d.cfg.Endpoint)
	req.Header.Set("x-amz-date", amzDate)
	req.Header.Set("x-amz-content-sha256", payloadHash)

	canonicalURI := fmt.Sprintf("/%s/%s", d.cfg.Bucket, cleanName)
	canonicalHeaders := fmt.Sprintf("host:%s\nx-amz-content-sha256:%s\nx-amz-date:%s\n",
		d.cfg.Endpoint, payloadHash, amzDate)
	signedHeaders := "host;x-amz-content-sha256;x-amz-date"

	canonicalRequest := fmt.Sprintf("%s\n%s\n%s\n%s\n%s\n%s",
		http.MethodDelete,
		canonicalURI,
		"",
		canonicalHeaders,
		signedHeaders,
		payloadHash,
	)

	hCanonical := sha256.Sum256([]byte(canonicalRequest))
	credentialScope := fmt.Sprintf("%s/%s/s3/aws4_request", dateStamp, d.cfg.Region)
	stringToSign := fmt.Sprintf("AWS4-HMAC-SHA256\n%s\n%s\n%s",
		amzDate,
		credentialScope,
		hex.EncodeToString(hCanonical[:]),
	)

	signingKey := getSignatureKey(d.cfg.SecretKey, dateStamp, d.cfg.Region, "s3")
	signature := hmacSHA256(signingKey, []byte(stringToSign))

	authHeader := fmt.Sprintf("AWS4-HMAC-SHA256 Credential=%s/%s, SignedHeaders=%s, Signature=%s",
		d.cfg.AccessKey, credentialScope, signedHeaders, hex.EncodeToString(signature))
	req.Header.Set("Authorization", authHeader)

	resp, err := d.client.Do(req)
	if err != nil {
		return fmt.Errorf("godeniter/storage: S3 删除请求失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("godeniter/storage: S3 删除失败，返回状态 %d", resp.StatusCode)
	}
	return nil
}

// AWS SigV4 纯标准库签名辅助
func hmacSHA256(key []byte, data []byte) []byte {
	h := hmac.New(sha256.New, key)
	h.Write(data)
	return h.Sum(nil)
}

func getSignatureKey(key, dateStamp, regionName, serviceName string) []byte {
	kDate := hmacSHA256([]byte("AWS4"+key), []byte(dateStamp))
	kRegion := hmacSHA256(kDate, []byte(regionName))
	kService := hmacSHA256(kRegion, []byte(serviceName))
	kSigning := hmacSHA256(kService, []byte("aws4_request"))
	return kSigning
}

// Copyright (c) 2026 xbt. All rights reserved.
// Godeniter is licensed under the GNU General Public License v3.0 (GPL-3.0).
// See LICENSE file in the project root for full license information.

package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"flag"
	"fmt"
	"math/big"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

func main() {
	var (
		name    string
		org     string
		country string
		years   int
		outDir  string
		pfxPass string
		genPfx  bool
	)

	flag.StringVar(&name, "name", "Godeniter Software Developer", "证书颁发者/签名者通用名称 (Common Name)")
	flag.StringVar(&org, "org", "Godeniter Project", "组织/公司名称 (Organization)")
	flag.StringVar(&country, "c", "CN", "两字母国家代码 (Country)")
	flag.IntVar(&years, "years", 5, "证书有效年限")
	flag.StringVar(&outDir, "out", "./certs", "证书输出目录")
	flag.StringVar(&pfxPass, "pass", "123456", "生成的 PFX/P12 文件加密密码")
	flag.BoolVar(&genPfx, "pfx", true, "若本地存在 openssl，是否自动合成 .pfx 签名包")
	flag.Parse()

	if err := os.MkdirAll(outDir, 0755); err != nil {
		fmt.Printf(">> [ERROR] 创建输出目录失败: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("==========================================================")
	fmt.Println(" 🔐 Godeniter 0 依赖 Windows 代码签名证书生成工具")
	fmt.Println("==========================================================")
	fmt.Printf(">> 正在基于纯 Go 标准库生成 2048 位 RSA 密钥对...\n")

	// 1. 生成 2048 位 RSA 密钥对
	privKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		fmt.Printf(">> [ERROR] 生成 RSA 私钥失败: %v\n", err)
		os.Exit(1)
	}

	// 2. 生成序列号
	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		serialNumber = big.NewInt(time.Now().UnixNano())
	}

	now := time.Now()
	notBefore := now.Add(-10 * time.Minute) // 提前10分钟，避免机器时钟微小偏差导致未生效
	notAfter := now.AddDate(years, 0, 0)

	// 3. 构建 X.509 证书模板 (核心：必须包含 ExtKeyUsageCodeSigning)
	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			CommonName:   name,
			Organization: []string{org},
			Country:      []string{country},
		},
		NotBefore: notBefore,
		NotAfter:  notAfter,

		// 密钥用法与代码签名专用扩展
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageCodeSigning},
		BasicConstraintsValid: true,
		IsCA:                  true, // 作为根证书自签名
	}

	// 4. 自签名生成 DER 编码的证书
	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &privKey.PublicKey, privKey)
	if err != nil {
		fmt.Printf(">> [ERROR] 生成 X.509 证书失败: %v\n", err)
		os.Exit(1)
	}

	// 5. 写入公钥证书 (DER 格式: .cer，Windows 资源管理器一键导入原生格式)
	cerPath := filepath.Join(outDir, "app_codesign.cer")
	if err := os.WriteFile(cerPath, derBytes, 0644); err != nil {
		fmt.Printf(">> [ERROR] 写入 .cer 公钥失败: %v\n", err)
		os.Exit(1)
	}

	// 6. 写入公钥证书 (PEM 格式: .crt)
	crtPath := filepath.Join(outDir, "app_codesign.crt")
	crtFile, err := os.Create(crtPath)
	if err != nil {
		fmt.Printf(">> [ERROR] 创建 .crt 证书文件失败: %v\n", err)
		os.Exit(1)
	}
	_ = pem.Encode(crtFile, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes})
	_ = crtFile.Close()

	// 7. 写入私钥 (PEM 格式: .key)
	keyPath := filepath.Join(outDir, "app_codesign.key")
	keyFile, err := os.Create(keyPath)
	if err != nil {
		fmt.Printf(">> [ERROR] 创建 .key 私钥文件失败: %v\n", err)
		os.Exit(1)
	}
	privDER := x509.MarshalPKCS1PrivateKey(privKey)
	_ = pem.Encode(keyFile, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: privDER})
	_ = keyFile.Close()

	fmt.Printf(">> [OK] 私钥生成成功: %s\n", keyPath)
	fmt.Printf(">> [OK] 公钥生成成功: %s (PEM) / %s (DER)\n", crtPath, cerPath)

	// 8. 若本地有 openssl，自动合成标准 PKCS#12 (.pfx)
	pfxPath := filepath.Join(outDir, "app_codesign.pfx")
	if genPfx {
		if opensslPath, err := exec.LookPath("openssl"); err == nil {
			cmd := exec.Command(opensslPath, "pkcs12", "-export",
				"-out", pfxPath,
				"-inkey", keyPath,
				"-in", crtPath,
				"-passout", "pass:"+pfxPass,
			)
			if out, cmdErr := cmd.CombinedOutput(); cmdErr == nil {
				fmt.Printf(">> [OK] 自动合成 Windows PFX 签名证书: %s (密码: %s)\n", pfxPath, pfxPass)
			} else {
				fmt.Printf(">> [NOTICE] openssl 转换 PFX 提示: %s\n", string(out))
			}
		}
	}

	fmt.Println("==========================================================")
	fmt.Println(" 🎉 证书生成完毕！使用说明：")
	fmt.Println("----------------------------------------------------------")
	fmt.Println("1. 【私钥 (app_codesign.key / .pfx)】：放在开发机保密！")
	fmt.Println("   用于执行签名：osslsigncode 或 Windows signtool 给 app.exe 盖章。")
	fmt.Println()
	fmt.Println("2. 【公钥 (app_codesign.cer)】：公开分发给客户电脑！")
	fmt.Println("   在 Windows 客户电脑上，以管理员权限执行一行命令即可永久信任：")
	fmt.Printf("   certutil -addstore -f \"ROOT\" %s\n", filepath.Base(cerPath))
	fmt.Println("   (导入后，客户机运行你签名的 app.exe 将彻底免除 SmartScreen 拦截警告！)")
	fmt.Println("==========================================================")
}

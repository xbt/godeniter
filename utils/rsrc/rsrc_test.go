// Copyright (c) 2026 xbt. All rights reserved.
// Godeniter is licensed under the GNU General Public License v3.0 (GPL-3.0).
// See LICENSE file in the project root for full license information.

package rsrc

import (
	"bytes"
	"debug/pe"
	"encoding/binary"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"testing"
)

// buildSampleICO 内存生成一个有效的 32x32 尺寸测试 ICO
func buildSampleICO() []byte {
	img := image.NewRGBA(image.Rect(0, 0, 32, 32))
	for y := 0; y < 32; y++ {
		for x := 0; x < 32; x++ {
			img.Set(x, y, color.RGBA{R: 2, G: 132, B: 199, A: 255})
		}
	}
	var pngBuf bytes.Buffer
	_ = png.Encode(&pngBuf, img)
	pngBytes := pngBuf.Bytes()

	var icoBuf bytes.Buffer
	// ICONDIR
	binary.Write(&icoBuf, binary.LittleEndian, uint16(0))
	binary.Write(&icoBuf, binary.LittleEndian, uint16(1))
	binary.Write(&icoBuf, binary.LittleEndian, uint16(1))

	// ICONDIRENTRY
	icoBuf.WriteByte(32) // Width
	icoBuf.WriteByte(32) // Height
	icoBuf.WriteByte(0)  // ColorCount
	icoBuf.WriteByte(0)  // Reserved
	binary.Write(&icoBuf, binary.LittleEndian, uint16(1))
	binary.Write(&icoBuf, binary.LittleEndian, uint16(32))
	binary.Write(&icoBuf, binary.LittleEndian, uint32(len(pngBytes)))
	binary.Write(&icoBuf, binary.LittleEndian, uint32(22)) // 6 + 16 = 22

	icoBuf.Write(pngBytes)
	return icoBuf.Bytes()
}

func TestConvert(t *testing.T) {
	icoData := buildSampleICO()
	sysoData, err := Convert(icoData)
	if err != nil {
		t.Fatalf("Convert 失败: %v", err)
	}

	if len(sysoData) < 20+40 {
		t.Fatalf("生成的 syso 文件过小: %d 字节", len(sysoData))
	}

	// 验证生成的文件能够被 Go debug/pe 正确解析
	tmpFile := filepath.Join(t.TempDir(), "test_resource.syso")
	if err := os.WriteFile(tmpFile, sysoData, 0644); err != nil {
		t.Fatal(err)
	}

	f, err := pe.Open(tmpFile)
	if err != nil {
		t.Fatalf("pe.Open 无法读取生成的 COFF 文件: %v", err)
	}
	defer f.Close()

	if len(f.Sections) != 1 {
		t.Fatalf("预期 1 个 section，实际: %d", len(f.Sections))
	}
	if f.Sections[0].Name != ".rsrc" {
		t.Errorf("预期 section 名字为 .rsrc，实际: %s", f.Sections[0].Name)
	}
	if f.Sections[0].NumberOfRelocations == 0 {
		t.Errorf("预期包含重定位项，实际为 0")
	}
}

func TestAutoDetectAndGenerate(t *testing.T) {
	tmpDir := t.TempDir()

	// 1. 无 app.ico 时，应返回 false
	detected, _, err := AutoDetectAndGenerate(tmpDir)
	if err != nil {
		t.Fatal(err)
	}
	if detected {
		t.Errorf("预期未检测到 app.ico")
	}

	// 2. 写入 app.ico 时，自动检测并生成 resource_windows_amd64.syso
	icoPath := filepath.Join(tmpDir, "app.ico")
	if err := os.WriteFile(icoPath, buildSampleICO(), 0644); err != nil {
		t.Fatal(err)
	}

	detected, sysoPath, err := AutoDetectAndGenerate(tmpDir)
	if err != nil {
		t.Fatalf("AutoDetectAndGenerate 失败: %v", err)
	}
	if !detected {
		t.Errorf("预期检测到 app.ico")
	}

	expectedSyso := filepath.Join(tmpDir, "resource_windows_amd64.syso")
	if sysoPath != expectedSyso {
		t.Errorf("预期输出文件为 %s，实际: %s", expectedSyso, sysoPath)
	}

	info, err := os.Stat(expectedSyso)
	if err != nil || info.Size() == 0 {
		t.Errorf("生成的 syso 文件不存在或大小为 0")
	}
}

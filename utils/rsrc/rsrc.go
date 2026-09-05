// Copyright (c) 2026 xbt. All rights reserved.
// Godeniter is licensed under the GNU General Public License v3.0 (GPL-3.0).
// See LICENSE file in the project root for full license information.

package rsrc

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// ICO 格式头部
type icoHeader struct {
	Reserved uint16
	Type     uint16
	Count    uint16
}

// ICO 每个图像帧目录条目 (16 字节)
type icoDirEntry struct {
	Width       uint8
	Height      uint8
	ColorCount  uint8
	Reserved    uint8
	Planes      uint16
	BitCount    uint16
	BytesInRes  uint32
	ImageOffset uint32
}

// RT_GROUP_ICON 中每个图像帧条目 (14 字节)
type grpIconDirEntry struct {
	Width      uint8
	Height     uint8
	ColorCount uint8
	Reserved   uint8
	Planes     uint16
	BitCount   uint16
	BytesInRes uint32
	ID         uint16
}

// Windows 资源目录表头 (16 字节)
type resDirHeader struct {
	Characteristics uint32
	TimeDateStamp   uint32
	MajorVersion    uint16
	MinorVersion    uint16
	NamedEntries    uint16
	IDEntries       uint16
}

// Windows 资源目录条目 (8 字节)
type resDirEntry struct {
	ID           uint32
	OffsetToData uint32
}

// Windows 资源数据描述项 (16 字节)
type resDataEntry struct {
	OffsetToData uint32
	Size         uint32
	CodePage     uint32
	Reserved     uint32
}

// COFF 文件头 (20 字节)
type coffHeader struct {
	Machine              uint16
	NumberOfSections     uint16
	TimeDateStamp        uint32
	PointerToSymbolTable uint32
	NumberOfSymbols      uint32
	SizeOfOptionalHeader uint16
	Characteristics      uint16
}

// COFF Section 头 (40 字节)
type coffSectionHeader struct {
	Name                 [8]byte
	VirtualSize          uint32
	VirtualAddress       uint32
	SizeOfRawData        uint32
	PointerToRawData     uint32
	PointerToRelocations uint32
	PointerToLinenumbers uint32
	NumberOfRelocations  uint16
	NumberOfLinenumbers  uint16
	Characteristics      uint32
}

// COFF 重定位记录 (10 字节)
type coffReloc struct {
	VirtualAddress   uint32
	SymbolTableIndex uint32
	Type             uint16
}

// COFF 符号表项 (18 字节)
type coffSymbol struct {
	Name               [8]byte
	Value              uint32
	SectionNumber      int16
	Type               uint16
	StorageClass       uint8
	NumberOfAuxSymbols uint8
}

const (
	rtIcon      = 3
	rtGroupIcon = 14

	imageFileMachineAMD64 = 0x8664
	imageRelAMD64Addr32NB = 0x0003 // 32-bit RVA without image base
)

// Convert 将 ICO 二进制字节转换为符合 Windows x64 COFF 规范的 .syso 资源目标文件。
// 全程采用 100% 纯 Go 标准库实现，零外部依赖。
func Convert(icoData []byte) ([]byte, error) {
	r := bytes.NewReader(icoData)
	var hdr icoHeader
	if err := binary.Read(r, binary.LittleEndian, &hdr); err != nil {
		return nil, fmt.Errorf("读取 ICO 头部失败: %w", err)
	}
	if hdr.Reserved != 0 || hdr.Type != 1 || hdr.Count == 0 {
		return nil, errors.New("无效的 ICO 文件格式 (Reserved 必须为 0 且 Type 必须为 1)")
	}

	entries := make([]icoDirEntry, hdr.Count)
	for i := 0; i < int(hdr.Count); i++ {
		if err := binary.Read(r, binary.LittleEndian, &entries[i]); err != nil {
			return nil, fmt.Errorf("读取 ICO 目录条目 [%d] 失败: %w", i, err)
		}
	}

	// 提取各个 ICON 图像帧数据，并组装 RT_GROUP_ICON 数据结构
	iconDataList := make([][]byte, hdr.Count)
	var grpBuf bytes.Buffer
	_ = binary.Write(&grpBuf, binary.LittleEndian, hdr.Reserved)
	_ = binary.Write(&grpBuf, binary.LittleEndian, hdr.Type)
	_ = binary.Write(&grpBuf, binary.LittleEndian, hdr.Count)

	for i, e := range entries {
		if int(e.ImageOffset+e.BytesInRes) > len(icoData) {
			return nil, fmt.Errorf("ICO 图像帧 [%d] 偏移越界", i)
		}
		iconDataList[i] = icoData[e.ImageOffset : e.ImageOffset+e.BytesInRes]

		grpEntry := grpIconDirEntry{
			Width:      e.Width,
			Height:     e.Height,
			ColorCount: e.ColorCount,
			Reserved:   e.Reserved,
			Planes:     e.Planes,
			BitCount:   e.BitCount,
			BytesInRes: e.BytesInRes,
			ID:         uint16(i + 1), // 对应 RT_ICON 的资源 ID，从 1 递增
		}
		if err := binary.Write(&grpBuf, binary.LittleEndian, grpEntry); err != nil {
			return nil, err
		}
	}
	groupIconData := grpBuf.Bytes()

	// 构建 Windows .rsrc 三级资源目录树
	count := int(hdr.Count)
	totalItems := count + 1 // Count 个单个图标 + 1 个图标组

	rootDirSize := 16 + 2*8
	iconTypeDirSize := 16 + count*8
	grpTypeDirSize := 16 + 1*8
	langDirSize := 16 + 1*8
	totalLangDirsSize := totalItems * langDirSize
	totalDataEntriesSize := totalItems * 16

	rootDirOffset := uint32(0)
	iconTypeDirOffset := rootDirOffset + uint32(rootDirSize)
	grpTypeDirOffset := iconTypeDirOffset + uint32(iconTypeDirSize)
	firstLangDirOffset := grpTypeDirOffset + uint32(grpTypeDirSize)
	firstDataEntryOffset := firstLangDirOffset + uint32(totalLangDirsSize)
	rawDataOffset := firstDataEntryOffset + uint32(totalDataEntriesSize)

	var rsrcBuf bytes.Buffer

	// 1. Root Dir (Type)
	_ = binary.Write(&rsrcBuf, binary.LittleEndian, resDirHeader{IDEntries: 2})
	_ = binary.Write(&rsrcBuf, binary.LittleEndian, resDirEntry{
		ID:           rtIcon,
		OffsetToData: iconTypeDirOffset | 0x80000000,
	})
	_ = binary.Write(&rsrcBuf, binary.LittleEndian, resDirEntry{
		ID:           rtGroupIcon,
		OffsetToData: grpTypeDirOffset | 0x80000000,
	})

	// 2. Type Dir (RT_ICON)
	_ = binary.Write(&rsrcBuf, binary.LittleEndian, resDirHeader{IDEntries: uint16(count)})
	for i := 0; i < count; i++ {
		langOffset := firstLangDirOffset + uint32(i*langDirSize)
		_ = binary.Write(&rsrcBuf, binary.LittleEndian, resDirEntry{
			ID:           uint32(i + 1),
			OffsetToData: langOffset | 0x80000000,
		})
	}

	// 3. Type Dir (RT_GROUP_ICON)
	_ = binary.Write(&rsrcBuf, binary.LittleEndian, resDirHeader{IDEntries: 1})
	grpLangOffset := firstLangDirOffset + uint32(count*langDirSize)
	_ = binary.Write(&rsrcBuf, binary.LittleEndian, resDirEntry{
		ID:           1, // 主图标组 ID 为 1
		OffsetToData: grpLangOffset | 0x80000000,
	})

	// 4. Lang Dirs (中性语言 0)
	for i := 0; i < totalItems; i++ {
		dataEntryOffset := firstDataEntryOffset + uint32(i*16)
		_ = binary.Write(&rsrcBuf, binary.LittleEndian, resDirHeader{IDEntries: 1})
		_ = binary.Write(&rsrcBuf, binary.LittleEndian, resDirEntry{
			ID:           0,
			OffsetToData: dataEntryOffset,
		})
	}

	// 5. Data Entries 与 Raw Data 追加
	var relocVirtualAddresses []uint32
	var rawDataBuf bytes.Buffer

	currDataOffset := rawDataOffset
	for i := 0; i < count; i++ {
		data := iconDataList[i]
		pad := (4 - (len(data) % 4)) % 4

		relocVirtualAddresses = append(relocVirtualAddresses, uint32(rsrcBuf.Len()))

		_ = binary.Write(&rsrcBuf, binary.LittleEndian, resDataEntry{
			OffsetToData: currDataOffset,
			Size:         uint32(len(data)),
		})

		rawDataBuf.Write(data)
		if pad > 0 {
			rawDataBuf.Write(make([]byte, pad))
		}
		currDataOffset += uint32(len(data) + pad)
	}

	// Group Icon Data Entry
	relocVirtualAddresses = append(relocVirtualAddresses, uint32(rsrcBuf.Len()))
	grpPad := (4 - (len(groupIconData) % 4)) % 4
	_ = binary.Write(&rsrcBuf, binary.LittleEndian, resDataEntry{
		OffsetToData: currDataOffset,
		Size:         uint32(len(groupIconData)),
	})
	rawDataBuf.Write(groupIconData)
	if grpPad > 0 {
		rawDataBuf.Write(make([]byte, grpPad))
	}

	rsrcBuf.Write(rawDataBuf.Bytes())
	rsrcData := rsrcBuf.Bytes()

	// 6. 组装为标准的 COFF 目标文件
	var coffBuf bytes.Buffer

	sizeOfHeaders := uint32(20 + 40)
	pointerToRawData := sizeOfHeaders
	sizeOfRawData := uint32(len(rsrcData))
	pointerToRelocs := pointerToRawData + sizeOfRawData
	numRelocs := uint16(len(relocVirtualAddresses))
	pointerToSymbols := pointerToRelocs + uint32(numRelocs*10)

	// COFF Header
	cHdr := coffHeader{
		Machine:              imageFileMachineAMD64,
		NumberOfSections:     1,
		TimeDateStamp:        0,
		PointerToSymbolTable: pointerToSymbols,
		NumberOfSymbols:      1,
		SizeOfOptionalHeader: 0,
		Characteristics:      0,
	}
	_ = binary.Write(&coffBuf, binary.LittleEndian, cHdr)

	// Section Header (.rsrc)
	var secName [8]byte
	copy(secName[:], []byte(".rsrc"))
	sHdr := coffSectionHeader{
		Name:                 secName,
		VirtualSize:          0,
		VirtualAddress:       0,
		SizeOfRawData:        sizeOfRawData,
		PointerToRawData:     pointerToRawData,
		PointerToRelocations: pointerToRelocs,
		PointerToLinenumbers: 0,
		NumberOfRelocations:  numRelocs,
		NumberOfLinenumbers:  0,
		Characteristics:      0x40000040, // IMAGE_SCN_CNT_INITIALIZED_DATA | IMAGE_SCN_MEM_READ
	}
	_ = binary.Write(&coffBuf, binary.LittleEndian, sHdr)

	// Section Data
	coffBuf.Write(rsrcData)

	// Relocations
	for _, vAddr := range relocVirtualAddresses {
		rel := coffReloc{
			VirtualAddress:   vAddr,
			SymbolTableIndex: 0, // 指向第 0 个符号 (.rsrc 段)
			Type:             imageRelAMD64Addr32NB,
		}
		_ = binary.Write(&coffBuf, binary.LittleEndian, rel)
	}

	// Symbol Table
	sym := coffSymbol{
		Name:          secName,
		Value:         0,
		SectionNumber: 1,
		Type:          0,
		StorageClass:  3, // IMAGE_SYM_CLASS_STATIC
	}
	_ = binary.Write(&coffBuf, binary.LittleEndian, sym)

	// String Table (空，固定 4 字节表示自身长度)
	_ = binary.Write(&coffBuf, binary.LittleEndian, uint32(4))

	return coffBuf.Bytes(), nil
}

// Generate 读取指定路径的 ICO 文件，转换并输出到指定的 .syso 文件路径。
func Generate(icoPath, outSysoPath string) error {
	icoBytes, err := os.ReadFile(icoPath)
	if err != nil {
		return fmt.Errorf("读取 ICO 文件 [%s] 失败: %w", icoPath, err)
	}

	sysoBytes, err := Convert(icoBytes)
	if err != nil {
		return fmt.Errorf("转换 ICO 为 SYSO 失败: %w", err)
	}

	_ = os.MkdirAll(filepath.Dir(outSysoPath), 0755)
	if err := os.WriteFile(outSysoPath, sysoBytes, 0644); err != nil {
		return fmt.Errorf("写入 SYSO 文件 [%s] 失败: %w", outSysoPath, err)
	}

	return nil
}

// AutoDetectAndGenerate 动态检测目录中的 app.ico 文件。
// 若检测到有效图标且资源文件需要更新，自动编译生成 resource_windows_amd64.syso。
// 返回值：
//   - detected: 是否检测到并成功处理了 app.ico
//   - sysoPath: 生成的 .syso 绝对/相对路径
//   - err: 发生的错误
func AutoDetectAndGenerate(dirs ...string) (bool, string, error) {
	targetDir := "."
	if len(dirs) > 0 && dirs[0] != "" {
		targetDir = dirs[0]
	}

	candidates := []string{
		filepath.Join(targetDir, "app.ico"),
		filepath.Join(targetDir, "favicon.ico"),
	}

	var foundICO string
	for _, cand := range candidates {
		if info, err := os.Stat(cand); err == nil && !info.IsDir() {
			foundICO = cand
			break
		}
	}

	if foundICO == "" {
		return false, "", nil
	}

	outSyso := filepath.Join(targetDir, "resource_windows_amd64.syso")

	// 检查时间戳：若 .syso 已存在且更新时间晚于 .ico，无需重复生成
	icoInfo, err := os.Stat(foundICO)
	if err == nil {
		if sysoInfo, err := os.Stat(outSyso); err == nil {
			if sysoInfo.ModTime().After(icoInfo.ModTime()) && sysoInfo.Size() > 0 {
				return true, outSyso, nil
			}
		}
	}

	if err := Generate(foundICO, outSyso); err != nil {
		return true, "", err
	}

	return true, outSyso, nil
}

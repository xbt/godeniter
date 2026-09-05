// Copyright (c) 2026 xbt. All rights reserved.
// Godeniter is licensed under the GNU General Public License v3.0 (GPL-3.0).
// See LICENSE file in the project root for full license information.

package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/xbt/godeniter/utils/rsrc"
)

func main() {
	icoPath := flag.String("ico", "app.ico", "要转换的 Windows .ico 图标路径")
	outPath := flag.String("o", "resource_windows_amd64.syso", "输出的 Windows COFF .syso 资源文件路径")
	autoMode := flag.Bool("auto", false, "动态自动探测当前目录下的 app.ico / favicon.ico 并按需编译")
	flag.Parse()

	if *autoMode {
		detected, sysoPath, err := rsrc.AutoDetectAndGenerate(".")
		if err != nil {
			fmt.Fprintf(os.Stderr, ">> [RSRC ERROR] 自动转换图标失败: %v\n", err)
			os.Exit(1)
		}
		if detected {
			fmt.Printf(">> [RSRC] 成功检测并处理 Windows 图标资源: %s\n", sysoPath)
		} else {
			fmt.Println(">> [RSRC] 未检测到 app.ico，跳过 Windows 资源编译。")
		}
		return
	}

	if _, err := os.Stat(*icoPath); os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, ">> [RSRC ERROR] 指定的 ICO 文件不存在: %s\n", *icoPath)
		os.Exit(1)
	}

	if err := rsrc.Generate(*icoPath, *outPath); err != nil {
		fmt.Fprintf(os.Stderr, ">> [RSRC ERROR] 转换失败: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf(">> [RSRC] 成功将 [%s] 转换为 Windows 资源文件 [%s]\n", *icoPath, *outPath)
}

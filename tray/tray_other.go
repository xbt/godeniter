//go:build !darwin && !windows

// Copyright (c) 2026 xbt. All rights reserved.
// Godeniter is licensed under the GNU General Public License v3.0 (GPL-3.0).
// See LICENSE file in the project root for full license information.

package tray

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"runtime"
	"syscall"
)

var quitChan = make(chan struct{})

// Run 在 Linux 或无图形界面环境下优雅降级为前台信号监听
func Run(opts Options) error {
	log.Printf(">> [Godeniter Tray] 当前系统环境 (%s) 不支持或未配置桌面托盘，已自动降级为信号监听模式\n", runtime.GOOS)
	log.Printf(">> [Godeniter Tray] 服务访问地址: %s\n", opts.URL)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	select {
	case <-sigChan:
		log.Println("\n>> [Godeniter Tray] 收到系统退出信号...")
	case <-quitChan:
		log.Println("\n>> [Godeniter Tray] 收到应用退出指令...")
	}

	if opts.OnExit != nil {
		opts.OnExit()
	}

	return nil
}

// Quit 退出降级监听循环
func Quit() {
	select {
	case quitChan <- struct{}{}:
	default:
	}
}

// ShowAbout 打印应用关于信息到控制台
func ShowAbout(opts Options) {
	appTitle := opts.Title
	if appTitle == "" {
		appTitle = "Godeniter"
	}
	version := opts.Version
	if version == "" {
		version = "v1.0.0"
	}
	fmt.Printf("\n--- 关于 %s ---\n版本: %s\n进程 PID: %d\n监听端口: %s\n底层框架: Godeniter (Go %s)\n---------------------\n\n",
		appTitle, version, os.Getpid(), opts.Port, runtime.Version())
}

// HideConsole 在非 Windows 环境下为空实现
func HideConsole() {}

// ShowConsole 在非 Windows 环境下为空实现
func ShowConsole() {}

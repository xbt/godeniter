//go:build darwin

package tray

/*
#cgo CFLAGS: -x objective-c
#cgo LDFLAGS: -framework Cocoa
#include "tray_darwin.h"
#include <stdlib.h>
*/
import "C"

import (
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"sync"
	"syscall"
	"unsafe"
)

var (
	cbLock    sync.RWMutex
	cbMap     = make(map[int]func())
	cbCounter = 1000
)

//export trayMenuItemCallback
func trayMenuItemCallback(id C.int) {
	cbLock.RLock()
	fn, exists := cbMap[int(id)]
	cbLock.RUnlock()

	if exists && fn != nil {
		go fn()
	}
}

// Run 启动 macOS 状态栏应用 (阻塞主线程并运行 Cocoa 事件循环)
func Run(opts Options) error {
	// macOS 严格要求所有 AppKit/Cocoa GUI 必须在主线程执行
	runtime.LockOSThread()

	// 1. 若配置了图标物理路径，自动读取内容
	if len(opts.IconBytes) == 0 && opts.IconPath != "" {
		if data, err := os.ReadFile(opts.IconPath); err == nil {
			opts.IconBytes = data
		}
	}

	// 2. 初始化 Cocoa NSApplication 为 Accessory 模式
	C.native_init_app()

	// 捕获系统退出信号 (Ctrl+C 与 kill)，确保在终端按 Ctrl+C 能平滑关闭并安全退出
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigChan
		if opts.OnExit != nil {
			opts.OnExit()
		}
		Quit()
		os.Exit(0)
	}()

	// 3. 创建状态栏图标与提示
	var cIcon unsafe.Pointer
	var cIconLen C.size_t
	if len(opts.IconBytes) > 0 {
		cIcon = unsafe.Pointer(&opts.IconBytes[0])
		cIconLen = C.size_t(len(opts.IconBytes))
	}

	appTitle := opts.Title
	if appTitle == "" {
		appTitle = "Godeniter"
	}
	cTitle := C.CString(appTitle)
	defer C.free(unsafe.Pointer(cTitle))

	tip := opts.Tooltip
	if tip == "" {
		tip = appTitle
	}
	cTooltip := C.CString(tip)
	defer C.free(unsafe.Pointer(cTooltip))

	C.native_create_status_bar(cIcon, cIconLen, cTitle, cTooltip)

	// 4. 构建菜单项清单 (整合用户扩展菜单与经典四件套)
	fullMenuItems := buildFullMenuItems(opts)

	cbLock.Lock()
	cbMap = make(map[int]func())
	cbCounter = 1000

	cItems := make([]C.TrayMenuItemC, len(fullMenuItems))
	for i, item := range fullMenuItems {
		if item.IsSeparator {
			cItems[i].is_separator = 1
			continue
		}

		cbCounter++
		currentID := cbCounter
		if item.OnClick != nil {
			cbMap[currentID] = item.OnClick
		}

		cItems[i].title = C.CString(item.Title)
		defer C.free(unsafe.Pointer(cItems[i].title))
		cItems[i].callback_id = C.int(currentID)
		if item.Disabled {
			cItems[i].disabled = 1
		}
		if item.Checked {
			cItems[i].checked = 1
		}
	}
	cbLock.Unlock()

	var pItems *C.TrayMenuItemC
	if len(cItems) > 0 {
		pItems = &cItems[0]
	}
	C.native_update_menu(pItems, C.int(len(cItems)))

	// 5. 启动原生主事件循环 (阻塞当前线程)
	C.native_run_loop()

	return nil
}

// Quit 退出状态栏主事件循环
func Quit() {
	C.native_quit_loop()
}

// ShowAbout 弹出关于对话框
func ShowAbout(opts Options) {
	appTitle := opts.Title
	if appTitle == "" {
		appTitle = "Godeniter"
	}
	version := opts.Version
	if version == "" {
		version = "v1.0.0"
	}
	msg := fmt.Sprintf("名称: %s\n版本: %s\n进程 PID: %d\n监听端口: %s\n底层框架: Godeniter (Go %s)",
		appTitle, version, os.Getpid(), opts.Port, runtime.Version())

	cTitle := C.CString("关于 " + appTitle)
	cMsg := C.CString(msg)
	defer C.free(unsafe.Pointer(cTitle))
	defer C.free(unsafe.Pointer(cMsg))

	C.native_show_alert(cTitle, cMsg)
}

// buildFullMenuItems 组装经典四件套与开发者自定义菜单
func buildFullMenuItems(opts Options) []MenuItem {
	var items []MenuItem

	// 1. 🌐 打开管理后台
	if opts.URL != "" {
		items = append(items, MenuItem{
			Title: "🌐 打开管理后台",
			OnClick: func() {
				_ = OpenURL(opts.URL)
			},
		})
	}

	// 2. 📁 打开应用目录
	appDir := opts.AppDir
	if appDir == "" {
		appDir = GetExecutableDir()
	}
	items = append(items, MenuItem{
		Title: "📁 打开应用目录",
		OnClick: func() {
			_ = OpenFolder(appDir)
		},
	})

	// 3. ℹ️ 关于系统
	items = append(items, MenuItem{
		Title: "ℹ️ 关于系统",
		OnClick: func() {
			ShowAbout(opts)
		},
	})

	// 4. 用户自定义菜单项
	if len(opts.Menus) > 0 {
		items = append(items, MenuItem{IsSeparator: true})
		items = append(items, opts.Menus...)
	}

	// 5. 分割线与退出
	items = append(items, MenuItem{IsSeparator: true})
	items = append(items, MenuItem{
		Title: "⏹️ 退出程序",
		OnClick: func() {
			if opts.OnExit != nil {
				opts.OnExit()
			}
			Quit()
			os.Exit(0)
		},
	})

	return items
}

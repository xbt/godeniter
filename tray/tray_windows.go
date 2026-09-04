//go:build windows

package tray

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"sync"
	"syscall"
	"unsafe"
)

var (
	moduser32   = syscall.NewLazyDLL("user32.dll")
	modshell32  = syscall.NewLazyDLL("shell32.dll")
	modkernel32 = syscall.NewLazyDLL("kernel32.dll")

	procRegisterClassExW        = moduser32.NewProc("RegisterClassExW")
	procCreateWindowExW         = moduser32.NewProc("CreateWindowExW")
	procDefWindowProcW          = moduser32.NewProc("DefWindowProcW")
	procDestroyWindow           = moduser32.NewProc("DestroyWindow")
	procPostQuitMessage         = moduser32.NewProc("PostQuitMessage")
	procPostMessageW            = moduser32.NewProc("PostMessageW")
	procGetMessageW             = moduser32.NewProc("GetMessageW")
	procTranslateMessage        = moduser32.NewProc("TranslateMessage")
	procDispatchMessageW        = moduser32.NewProc("DispatchMessageW")
	procLoadIconW               = moduser32.NewProc("LoadIconW")
	procCreatePopupMenu         = moduser32.NewProc("CreatePopupMenu")
	procAppendMenuW             = moduser32.NewProc("AppendMenuW")
	procTrackPopupMenu          = moduser32.NewProc("TrackPopupMenu")
	procDestroyMenu             = moduser32.NewProc("DestroyMenu")
	procGetCursorPos            = moduser32.NewProc("GetCursorPos")
	procSetForegroundWindow     = moduser32.NewProc("SetForegroundWindow")
	procMessageBoxW             = moduser32.NewProc("MessageBoxW")
	procCreateIconFromResourceEx = moduser32.NewProc("CreateIconFromResourceEx")

	procShell_NotifyIconW       = modshell32.NewProc("Shell_NotifyIconW")
	procGetModuleHandleW        = modkernel32.NewProc("GetModuleHandleW")
)

const (
	WM_DESTROY       = 0x0002
	WM_USER          = 0x0400
	WM_TRAY_CALLBACK = WM_USER + 1024

	WM_LBUTTONDBLCLK = 0x0203
	WM_RBUTTONUP     = 0x0205
	WM_CONTEXTMENU   = 0x007B

	NIM_ADD    = 0x00000000
	NIM_MODIFY = 0x00000001
	NIM_DELETE = 0x00000002

	NIF_MESSAGE = 0x00000001
	NIF_ICON    = 0x00000002
	NIF_TIP     = 0x00000004

	MF_STRING    = 0x00000000
	MF_GRAYED    = 0x00000001
	MF_DISABLED  = 0x00000002
	MF_CHECKED   = 0x00000008
	MF_SEPARATOR = 0x00000800

	TPM_RIGHTALIGN  = 0x0008
	TPM_BOTTOMALIGN = 0x0020
	TPM_RETURNCMD   = 0x0100
	TPM_NONOTIFY    = 0x0080

	IMAGE_ICON      = 1
	LR_DEFAULTCOLOR = 0x0000
	IDI_APPLICATION = 32512
	MB_ICONINFO     = 0x00000040
	MB_OK           = 0x00000000
)

type POINT struct {
	X int32
	Y int32
}

type MSG struct {
	Hwnd    uintptr
	Message uint32
	WParam  uintptr
	LParam  uintptr
	Time    uint32
	Pt      POINT
}

type WNDCLASSEXW struct {
	CbSize        uint32
	Style         uint32
	LpfnWndProc   uintptr
	CbClsExtra    int32
	CbWndExtra    int32
	HInstance     uintptr
	HIcon         uintptr
	HCursor       uintptr
	HbrBackground uintptr
	LpszMenuName  *uint16
	LpszClassName *uint16
	HIconSm       uintptr
}

type NOTIFYICONDATAW struct {
	CbSize           uint32
	HWnd             uintptr
	UID              uint32
	UFlags           uint32
	UCallbackMessage uint32
	HIcon            uintptr
	SzTip            [128]uint16
	DwState          uint32
	DwStateMask      uint32
	SzInfo           [256]uint16
	TimeoutOrVersion uint32
	SzInfoTitle      [64]uint16
	DwInfoFlags      uint32
	GuidItem         [16]byte
	HBalloonIcon     uintptr
}

var (
	globalHWnd     uintptr
	globalNID      NOTIFYICONDATAW
	globalOpts     Options
	globalLock     sync.RWMutex
	winMenuActions = make(map[uintptr]func())
)

// wndProc 纯 Go 回调函数，接管 Win32 消息循环事件
func wndProc(hwnd uintptr, msg uint32, wParam, lParam uintptr) uintptr {
	switch msg {
	case WM_TRAY_CALLBACK:
		switch lParam {
		case WM_RBUTTONUP, WM_CONTEXTMENU:
			showWindowsTrayMenu(hwnd)
		case WM_LBUTTONDBLCLK:
			// 双击托盘图标: 默认在浏览器打开后台
			if globalOpts.URL != "" {
				_ = OpenURL(globalOpts.URL)
			}
		}
		return 0
	case WM_DESTROY:
		procPostQuitMessage.Call(0)
		return 0
	}

	ret, _, _ := procDefWindowProcW.Call(hwnd, uintptr(msg), wParam, lParam)
	return ret
}

// Run 启动 Windows 系统托盘程序 (纯 Go 标准库 syscall 驱动，0 CGO 依赖)
func Run(opts Options) error {
	runtime.LockOSThread()
	globalOpts = opts

	// 捕获系统退出信号 (Ctrl+C 与 kill)，确保控制台随时按 Ctrl+C 能平滑关闭并安全退出
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

	// 1. 若配置了本地图标物理路径，尝试读取
	if len(opts.IconBytes) == 0 && opts.IconPath != "" {
		if data, err := os.ReadFile(opts.IconPath); err == nil {
			opts.IconBytes = data
		}
	}

	hInstance, _, _ := procGetModuleHandleW.Call(0)
	className := syscall.StringToUTF16Ptr("GodeniterTrayWindowClass")

	// 2. 注册不可见的消息监听窗口类
	var wc WNDCLASSEXW
	wc.CbSize = uint32(unsafe.Sizeof(wc))
	wc.LpfnWndProc = syscall.NewCallback(wndProc)
	wc.HInstance = hInstance
	wc.LpszClassName = className
	procRegisterClassExW.Call(uintptr(unsafe.Pointer(&wc)))

	// 3. 创建消息处理窗口
	hwnd, _, _ := procCreateWindowExW.Call(
		0,
		uintptr(unsafe.Pointer(className)),
		uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr("GodeniterTray"))),
		0, 0, 0, 0, 0,
		0, 0, hInstance, 0,
	)
	if hwnd == 0 {
		return fmt.Errorf("创建托盘监听窗口失败")
	}
	globalHWnd = hwnd

	// 4. 加载托盘图标 (优先从内置/传入的 ICO 字节流动态创建，若无则使用系统默认图标)
	hIcon := loadIconFromBytes(opts.IconBytes)
	if hIcon == 0 {
		hIconRet, _, _ := procLoadIconW.Call(0, uintptr(IDI_APPLICATION))
		hIcon = hIconRet
	}

	// 5. 初始化托盘通知数据结构并注册到系统托盘
	var nid NOTIFYICONDATAW
	nid.CbSize = uint32(unsafe.Sizeof(nid))
	nid.HWnd = hwnd
	nid.UID = 1001
	nid.UFlags = NIF_MESSAGE | NIF_ICON | NIF_TIP
	nid.UCallbackMessage = WM_TRAY_CALLBACK
	nid.HIcon = hIcon

	tip := opts.Tooltip
	if tip == "" {
		tip = opts.Title
	}
	copyUTF16(nid.SzTip[:], tip)

	globalNID = nid
	ret, _, _ := procShell_NotifyIconW.Call(NIM_ADD, uintptr(unsafe.Pointer(&nid)))
	if ret == 0 {
		return fmt.Errorf("添加系统托盘图标失败")
	}

	// 6. 运行 Windows 原生消息队列循环 (阻塞当前线程)
	var m MSG
	for {
		r, _, _ := procGetMessageW.Call(uintptr(unsafe.Pointer(&m)), 0, 0, 0)
		if int32(r) <= 0 {
			break
		}
		procTranslateMessage.Call(uintptr(unsafe.Pointer(&m)))
		procDispatchMessageW.Call(uintptr(unsafe.Pointer(&m)))
	}

	return nil
}

// Quit 退出托盘并清理 Windows 图标与资源
func Quit() {
	if globalHWnd != 0 {
		procShell_NotifyIconW.Call(NIM_DELETE, uintptr(unsafe.Pointer(&globalNID)))
		procDestroyWindow.Call(globalHWnd)
		procPostQuitMessage.Call(0)
		globalHWnd = 0
	}
}

// ShowAbout 弹出 Windows 原生消息弹窗
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

	procMessageBoxW.Call(
		globalHWnd,
		uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(msg))),
		uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr("关于 "+appTitle))),
		MB_ICONINFO|MB_OK,
	)
}

// showWindowsTrayMenu 弹出右键菜单
func showWindowsTrayMenu(hwnd uintptr) {
	hMenu, _, _ := procCreatePopupMenu.Call()
	if hMenu == 0 {
		return
	}
	defer procDestroyMenu.Call(hMenu)

	items := buildFullMenuItems(globalOpts)

	globalLock.Lock()
	winMenuActions = make(map[uintptr]func())
	var menuIDCounter uintptr = 10000

	for _, it := range items {
		if it.IsSeparator {
			procAppendMenuW.Call(hMenu, MF_SEPARATOR, 0, 0)
			continue
		}

		menuIDCounter++
		currentID := menuIDCounter
		if it.OnClick != nil {
			winMenuActions[currentID] = it.OnClick
		}

		var flags uintptr = MF_STRING
		if it.Disabled {
			flags |= MF_GRAYED | MF_DISABLED
		}
		if it.Checked {
			flags |= MF_CHECKED
		}

		titlePtr := syscall.StringToUTF16Ptr(it.Title)
		procAppendMenuW.Call(hMenu, flags, currentID, uintptr(unsafe.Pointer(titlePtr)))
	}
	globalLock.Unlock()

	var pt POINT
	procGetCursorPos.Call(uintptr(unsafe.Pointer(&pt)))

	// 必须先置前当前窗口，否则菜单在点击空白区域时无法自动收起
	procSetForegroundWindow.Call(hwnd)

	// 阻塞弹出菜单并直接返回选中的菜单 ID
	cmd, _, _ := procTrackPopupMenu.Call(
		hMenu,
		TPM_RETURNCMD|TPM_NONOTIFY|TPM_BOTTOMALIGN|TPM_RIGHTALIGN,
		uintptr(pt.X),
		uintptr(pt.Y),
		0,
		hwnd,
		0,
	)

	// 微软官方规范: 必须在 TrackPopupMenu 之后投递 WM_NULL 消息，以确保点击空白区域收起后能再次正常右键弹出
	procPostMessageW.Call(hwnd, 0, 0, 0)

	// 后续处理被选中的动作
	if cmd > 0 {
		globalLock.RLock()
		action, exists := winMenuActions[cmd]
		globalLock.RUnlock()
		if exists && action != nil {
			go action()
		}
	}
}

// buildFullMenuItems 组装 Windows 菜单项 (经典四件套与扩展菜单)
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

	// 4. 自定义扩展菜单
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

// loadIconFromBytes 从 ICO 字节流中自动解析并提取图标创建 HICON 句柄
func loadIconFromBytes(data []byte) uintptr {
	if len(data) < 6 {
		return 0
	}

	// 检查 ICO 头部 (0x0000 0x0001 count)
	var reserved, imgType, count uint16
	r := bytes.NewReader(data)
	_ = binary.Read(r, binary.LittleEndian, &reserved)
	_ = binary.Read(r, binary.LittleEndian, &imgType)
	_ = binary.Read(r, binary.LittleEndian, &count)

	if reserved != 0 || imgType != 1 || count == 0 {
		// 若不是标准 ICO，尝试作为单图片资源直接交由 CreateIconFromResourceEx
		hIcon, _, _ := procCreateIconFromResourceEx.Call(
			uintptr(unsafe.Pointer(&data[0])),
			uintptr(len(data)),
			1, // TRUE 表示图标
			0x00030000,
			16, 16, // 托盘推荐标准 16x16 尺寸
			LR_DEFAULTCOLOR,
		)
		return hIcon
	}

	// 寻找最适合 16x16 或 32x32 的图标资源条目
	bestOffset := uint32(0)
	bestSize := uint32(0)
	var targetDiff int = 999

	for i := 0; i < int(count); i++ {
		var width, height, colors, res byte
		var planes, bpp uint16
		var size, offset uint32

		_ = binary.Read(r, binary.LittleEndian, &width)
		_ = binary.Read(r, binary.LittleEndian, &height)
		_ = binary.Read(r, binary.LittleEndian, &colors)
		_ = binary.Read(r, binary.LittleEndian, &res)
		_ = binary.Read(r, binary.LittleEndian, &planes)
		_ = binary.Read(r, binary.LittleEndian, &bpp)
		_ = binary.Read(r, binary.LittleEndian, &size)
		_ = binary.Read(r, binary.LittleEndian, &offset)

		w := int(width)
		if w == 0 {
			w = 256
		}
		diff := w - 16
		if diff < 0 {
			diff = -diff
		}
		if diff < targetDiff {
			targetDiff = diff
			bestOffset = offset
			bestSize = size
		}
	}

	if bestOffset > 0 && bestSize > 0 && int(bestOffset+bestSize) <= len(data) {
		resData := data[bestOffset : bestOffset+bestSize]
		hIcon, _, _ := procCreateIconFromResourceEx.Call(
			uintptr(unsafe.Pointer(&resData[0])),
			uintptr(bestSize),
			1, // TRUE 表示图标
			0x00030000,
			16, 16,
			LR_DEFAULTCOLOR,
		)
		return hIcon
	}

	return 0
}

func copyUTF16(dst []uint16, src string) {
	u16 := syscall.StringToUTF16(src)
	copy(dst, u16)
	if len(u16) < len(dst) {
		dst[len(u16)] = 0
	} else {
		dst[len(dst)-1] = 0
	}
}

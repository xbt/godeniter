package tray

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

// MenuItem 托盘菜单项定义
type MenuItem struct {
	Title       string     // 菜单项文本标题 (如 "打开管理后台")
	Tooltip     string     // 提示信息 (可选)
	Disabled    bool       // 是否置灰禁用
	Checked     bool       // 是否勾选状态
	IsSeparator bool       // 是否为分割线 (如果为 true，Title 忽略)
	OnClick     func()     // 点击事件回调函数
	SubMenus    []MenuItem // 二级子菜单 (可选)
}

// Options 托盘启动配置项
type Options struct {
	Title     string     // 应用名称 (如 "Godeniter Starter")
	Tooltip   string     // 鼠标悬停托盘图标时的提示文本
	IconBytes []byte     // 托盘图标二进制字节 (Windows 为 .ico 格式，macOS 支持 png/ico)
	IconPath  string     // 托盘图标文件物理路径 (优先于 IconBytes)
	URL       string     // Web 服务后台网址 (如 "http://127.0.0.1:8080")
	AppDir    string     // 应用根目录 (默认当前可执行程序所在目录)
	Version   string     // 应用版本 (如 "v1.0.0")
	Port      string     // 监听端口号 (如 ":8080")
	OnExit      func()     // 托盘退出前执行的清理或优雅停机钩子
	Menus       []MenuItem // 自定义扩展菜单项 (将与默认菜单合并)
	HideConsole bool       // 是否在启动托盘时自动隐藏控制台黑框 (Windows 原生生效，默认开启)
}

// DefaultOptions 返回开箱即用的托盘默认配置
func DefaultOptions() Options {
	dir, err := os.Getwd()
	if err != nil {
		dir = "."
	}
	return Options{
		Title:       "Godeniter",
		Tooltip:     "Godeniter Web Service",
		URL:         "http://127.0.0.1:8080",
		AppDir:      dir,
		Version:     "v1.0.0",
		HideConsole: true,
	}
}

// OpenURL 跨平台自动调起系统默认浏览器打开指定网址
func OpenURL(targetURL string) error {
	if targetURL == "" {
		return fmt.Errorf("url 不能为空")
	}

	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", targetURL)
	case "windows":
		// 使用 rundll32 调起默认浏览器关联，兼容所有 Windows 版本
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", targetURL)
	case "linux":
		cmd = exec.Command("xdg-open", targetURL)
	default:
		return fmt.Errorf("不支持的操作系统: %s", runtime.GOOS)
	}

	return cmd.Start()
}

// OpenFolder 跨平台自动调起系统文件管理器 (Finder / 资源管理器) 打开指定目录
func OpenFolder(folderPath string) error {
	if folderPath == "" {
		folderPath = "."
	}

	absPath, err := filepath.Abs(folderPath)
	if err != nil {
		absPath = folderPath
	}

	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", absPath)
	case "windows":
		cmd = exec.Command("explorer", absPath)
	case "linux":
		cmd = exec.Command("xdg-open", absPath)
	default:
		return fmt.Errorf("不支持的操作系统: %s", runtime.GOOS)
	}

	return cmd.Start()
}

// GetExecutableDir 获取当前可执行文件所在的绝对目录
func GetExecutableDir() string {
	exePath, err := os.Executable()
	if err != nil {
		dir, _ := os.Getwd()
		return dir
	}
	return filepath.Dir(exePath)
}

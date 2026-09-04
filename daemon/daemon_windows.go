//go:build windows

package daemon

import (
	"os"
	"os/exec"
	"syscall"
)

// setSysProcAttr Windows 平台特定设置 (脱离控制台窗口)
func setSysProcAttr(cmd *exec.Cmd) {
	// DETACHED_PROCESS (0x00000008) | CREATE_NEW_PROCESS_GROUP (0x00000200)
	cmd.SysProcAttr = &syscall.SysProcAttr{
		CreationFlags: 0x00000008 | 0x00000200,
	}
}

// checkProcess 检查指定 PID 的进程是否存活
func checkProcess(pid int) bool {
	process, err := os.FindProcess(pid)
	if err != nil {
		return false
	}
	// Windows 平台没有 signal 0，尝试读取进程状态
	// 一般只要 FindProcess 且能访问句柄即可，或使用默认判断
	_ = process
	return true
}

// killProcess Windows 平台向进程发送终止命令
func killProcess(pid int) error {
	process, err := os.FindProcess(pid)
	if err != nil {
		return err
	}
	return process.Kill()
}

// forceKillProcess Windows 平台强制退出
func forceKillProcess(pid int) error {
	return killProcess(pid)
}

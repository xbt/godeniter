//go:build !windows

// Copyright (c) 2026 xbt. All rights reserved.
// Godeniter is licensed under the GNU General Public License v3.0 (GPL-3.0).
// See LICENSE file in the project root for full license information.

package daemon

import (
	"os"
	"os/exec"
	"syscall"
)

// setSysProcAttr 设置 Unix/Linux/macOS 特定的进程属性 (脱离终端控制会话 Setsid)
func setSysProcAttr(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setsid: true, // 创建新的会话组，彻底脱离父终端会话
	}
}

// checkProcess 检查指定 PID 的进程是否存活
func checkProcess(pid int) bool {
	process, err := os.FindProcess(pid)
	if err != nil {
		return false
	}
	// 在 Unix 下向进程发送 signal 0 用于安全探测其存活性
	err = process.Signal(syscall.Signal(0))
	return err == nil
}

// killProcess 向进程发送 SIGTERM 优雅退出信号
func killProcess(pid int) error {
	process, err := os.FindProcess(pid)
	if err != nil {
		return err
	}
	return process.Signal(syscall.SIGTERM)
}

// forceKillProcess 向进程发送 SIGKILL 强制终结
func forceKillProcess(pid int) error {
	process, err := os.FindProcess(pid)
	if err != nil {
		return err
	}
	return process.Signal(syscall.SIGKILL)
}

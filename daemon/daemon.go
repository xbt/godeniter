// Copyright (c) 2026 xbt. All rights reserved.
// Godeniter is licensed under the GNU General Public License v3.0 (GPL-3.0).
// See LICENSE file in the project root for full license information.

package daemon

import (
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

)

const daemonWorkerEnv = "__GODENITER_DAEMON_WORKER__"

// Runner 能够被守护进程管理器托管启动的接口抽象 (如 *godeniter.Engine)
type Runner interface {
	Run(addr ...string) error
}


// Config 守护进程与服务生命周期配置项
type Config struct {
	PIDFile string `json:"pid_file"` // PID 记录文件存储路径 (默认: "./app.pid")
	LogFile string `json:"log_file"` // 后台运行日志重定向路径 (默认: "./app.log")
	Daemon  bool   `json:"daemon"`   // 是否启用守护进程模式 (默认: false 前台运行)
}

// DefaultConfig 返回开箱即用的守护进程默认配置
func DefaultConfig() Config {
	return Config{
		PIDFile: "./app.pid",
		LogFile: "./app.log",
		Daemon:  false,
	}
}

// Manager 守护进程与服务生命周期管理器
type Manager struct {
	cfg Config
}

// New 实例化管理器
func New(cfg Config) *Manager {
	if cfg.PIDFile == "" {
		cfg.PIDFile = "./app.pid"
	}
	if cfg.LogFile == "" {
		cfg.LogFile = "./app.log"
	}
	return &Manager{cfg: cfg}
}

// Run 统一接管服务启动与命令行生命周期指令
func Run(app Runner, addr string, cfgs ...Config) error {
	cfg := DefaultConfig()
	if len(cfgs) > 0 {
		cfg = cfgs[0]
		if cfg.PIDFile == "" {
			cfg.PIDFile = "./app.pid"
		}
		if cfg.LogFile == "" {
			cfg.LogFile = "./app.log"
		}
	}

	m := New(cfg)

	// 1. 若当前处于守护子进程 (Worker) 模式，记录 PID 并阻塞启动服务
	if os.Getenv(daemonWorkerEnv) == "1" {
		return m.runWorker(app, addr)
	}

	// 2. 检查命令行参数中是否包含生命周期指令
	cmd := ""
	if len(os.Args) > 1 {
		arg := strings.ToLower(os.Args[1])
		switch arg {
		case "start", "stop", "restart", "status":
			cmd = arg
		}
	}

	switch cmd {
	case "start":
		if err := m.Start(addr); err != nil {
			fmt.Printf(">> [ERROR] 启动守护进程失败: %v\n", err)
			os.Exit(1)
		}
		os.Exit(0)
		return nil
	case "stop":
		if err := m.Stop(); err != nil {
			fmt.Printf(">> [ERROR] 停止服务失败: %v\n", err)
			os.Exit(1)
		}
		os.Exit(0)
		return nil
	case "restart":
		if err := m.Restart(addr); err != nil {
			fmt.Printf(">> [ERROR] 重启服务失败: %v\n", err)
			os.Exit(1)
		}
		os.Exit(0)
		return nil
	case "status":
		m.Status(addr)
		os.Exit(0)
		return nil
	default:
		// 未指定子命令
		if cfg.Daemon {
			// 若配置文件指定了 daemon: true，自动后台启动
			return m.Start(addr)
		}
		// 默认前台开发调试运行
		return app.Run(addr)
	}
}

// runWorker 守护子进程内部执行逻辑
func (m *Manager) runWorker(app Runner, addr string) error {
	// 将自身 PID 写入文件
	pid := os.Getpid()
	_ = os.MkdirAll(filepath.Dir(m.cfg.PIDFile), 0755)
	_ = os.WriteFile(m.cfg.PIDFile, []byte(strconv.Itoa(pid)), 0644)

	// 退出时清理 PID 文件
	defer func() {
		_ = os.Remove(m.cfg.PIDFile)
	}()

	return app.Run(addr)
}

// Start 后台启动服务并返回命令行
func (m *Manager) Start(addr string) error {
	// 检查现有进程是否已经在运行中
	if pid, ok := m.readPID(); ok && checkProcess(pid) {
		fmt.Println("==========================================================")
		fmt.Printf(" >> [WARN] Godeniter 服务已经在后台运行中！(PID: %d)\n", pid)
		fmt.Printf(" >> 监听端口:    %s\n", addr)
		printServerURLs(addr)
		printOpsCommands(m.cfg.LogFile)
		return nil
	}

	// 1. 预检端口可用性，防止后台子进程静默因端口占用而失败
	if addr != "" {
		testLn, err := net.Listen("tcp", addr)
		if err != nil {
			return fmt.Errorf("端口 [%s] 已被占用，无法在后台启动服务！请执行 'lsof -ti %s | xargs kill -9' 释放端口，或在 config.json 中更换端口", addr, addr)
		}
		_ = testLn.Close()
	}

	// 准备日志文件目录
	_ = os.MkdirAll(filepath.Dir(m.cfg.LogFile), 0755)
	logFile, err := os.OpenFile(m.cfg.LogFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("创建日志文件 [%s] 失败: %w", m.cfg.LogFile, err)
	}
	defer logFile.Close()

	// 提取可执行程序路径与参数 (剔除 start/restart 等命令，避免死循环)
	binPath, err := os.Executable()
	if err != nil {
		binPath = os.Args[0]
	}

	var childArgs []string
	for _, arg := range os.Args[1:] {
		low := strings.ToLower(arg)
		if low != "start" && low != "restart" {
			childArgs = append(childArgs, arg)
		}
	}

	cmd := exec.Command(binPath, childArgs...)
	cmd.Env = append(os.Environ(), daemonWorkerEnv+"=1")
	cmd.Stdout = logFile
	cmd.Stderr = logFile
	cmd.Stdin = nil

	// 平台特定的进程属性 (脱离终端会话)
	setSysProcAttr(cmd)

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("启动后台进程失败: %w", err)
	}

	pid := cmd.Process.Pid
	// 先将子进程 PID 暂存到 PID 文件
	_ = os.WriteFile(m.cfg.PIDFile, []byte(strconv.Itoa(pid)), 0644)

	// 等待 250ms 验证子进程是否正常存活
	time.Sleep(250 * time.Millisecond)
	if !checkProcess(pid) {
		_ = os.Remove(m.cfg.PIDFile)
		return fmt.Errorf("子进程启动后立即异常退出，请查看日志: %s", m.cfg.LogFile)
	}

	fmt.Println("==========================================================")
	fmt.Printf(" >> Godeniter 2.0 服务已成功在后台启动 (Daemon Mode)!\n")
	fmt.Printf(" >> 进程 PID:    %d (已写入 %s)\n", pid, m.cfg.PIDFile)
	fmt.Printf(" >> 监听端口:    %s\n", addr)
	printServerURLs(addr)
	fmt.Printf(" >> 输出日志:    %s\n", m.cfg.LogFile)
	printOpsCommands(m.cfg.LogFile)

	return nil
}

// Stop 优雅停止后台服务
func (m *Manager) Stop() error {
	cmd := getExecCommand()
	pid, ok := m.readPID()
	if !ok {
		fmt.Printf(">> [STATUS] 服务未运行 (未检测到 PID 文件: %s)\n", m.cfg.PIDFile)
		fmt.Printf(">> [TIP] 启动服务指令: %s start\n", cmd)
		return nil
	}

	if !checkProcess(pid) {
		_ = os.Remove(m.cfg.PIDFile)
		fmt.Printf(">> [STATUS] 目标进程 (PID: %d) 已不存在，已自动清理失效的 PID 文件。\n", pid)
		fmt.Printf(">> [TIP] 启动服务指令: %s start\n", cmd)
		return nil
	}

	fmt.Printf(">> [STOP] 正在向服务发送安全退出信号 (PID: %d)...\n", pid)
	if err := killProcess(pid); err != nil {
		return fmt.Errorf("发送退出信号失败: %w", err)
	}

	// 轮询等待最多 5 秒，等待其平滑停机与连接释放
	for i := 0; i < 25; i++ {
		time.Sleep(200 * time.Millisecond)
		if !checkProcess(pid) {
			_ = os.Remove(m.cfg.PIDFile)
			fmt.Printf(">> [STOP] Godeniter 服务已安全优雅退出！\n")
			fmt.Printf(">> [TIP] 如需重新启动服务，请执行: %s start\n", cmd)
			return nil
		}
	}

	// 若超时仍未退出，尝试强制清理
	_ = forceKillProcess(pid)
	_ = os.Remove(m.cfg.PIDFile)
	fmt.Printf(">> [WARN] 服务在超时时间内未自行退出，已执行终结。\n")
	fmt.Printf(">> [TIP] 如需重新启动服务，请执行: %s start\n", cmd)
	return nil
}

// Restart 优雅重启服务
func (m *Manager) Restart(addr string) error {
	fmt.Println(">> [RESTART] 正在平滑重启服务...")
	_ = m.Stop()
	time.Sleep(500 * time.Millisecond)
	return m.Start(addr)
}

// Status 检查当前服务运行状态
func (m *Manager) Status(addr ...string) {
	listenAddr := ":8080"
	if len(addr) > 0 && addr[0] != "" {
		listenAddr = addr[0]
	}

	cmd := getExecCommand()
	pid, ok := m.readPID()
	if !ok {
		fmt.Println("==========================================================")
		fmt.Printf(" >> Godeniter 服务状态: [未运行 ⚪] (PID 文件不存在)\n")
		fmt.Printf(" >> 启动指令:    %s start\n", cmd)
		fmt.Println("==========================================================")
		return
	}

	if !checkProcess(pid) {
		_ = os.Remove(m.cfg.PIDFile)
		fmt.Println("==========================================================")
		fmt.Printf(" >> Godeniter 服务状态: [已停止 ⚪] (残留 PID 文件已自动清理)\n")
		fmt.Printf(" >> 启动指令:    %s start\n", cmd)
		fmt.Println("==========================================================")
		return
	}

	fmt.Printf("==========================================================\n")
	fmt.Printf(" >> Godeniter 服务状态: [运行中 🟢]\n")
	fmt.Printf(" >> 运行 PID:    %d (存活)\n", pid)
	fmt.Printf(" >> 监听端口:    %s\n", listenAddr)
	printServerURLs(listenAddr)
	fmt.Printf(" >> PID 文件:    %s\n", m.cfg.PIDFile)
	fmt.Printf(" >> 日志文件:    %s\n", m.cfg.LogFile)
	printOpsCommands(m.cfg.LogFile)
}

// getExecCommand 智能解析当前执行的命令原型（源码开发模式或二进制调用路径）
func getExecCommand() string {
	arg0 := os.Args[0]
	base := filepath.Base(arg0)

	// 判断是否为 go run 临时执行环境
	if strings.Contains(arg0, "go-build") || strings.Contains(arg0, "go-cache") || strings.Contains(arg0, "/b001/exe/") || base == "main" || base == "main.exe" {
		if _, err := os.Stat("main.go"); err == nil {
			return "go run main.go"
		}
	}

	// 二进制方式运行
	// 如果是相对路径，确保带上 ./ 前缀以防直接复制报 command not found
	if !filepath.IsAbs(arg0) {
		if !strings.HasPrefix(arg0, "."+string(filepath.Separator)) && !strings.HasPrefix(arg0, "./") {
			return "./" + arg0
		}
		return arg0
	}

	// 如果传入的是绝对路径，若在当前工作目录下，转换为更简洁直观的相对路径
	if wd, err := os.Getwd(); err == nil {
		if rel, err := filepath.Rel(wd, arg0); err == nil && !strings.HasPrefix(rel, "..") {
			if !strings.HasPrefix(rel, "."+string(filepath.Separator)) && !strings.HasPrefix(rel, "./") {
				return "./" + rel
			}
			return rel
		}
	}

	return arg0
}

// parsePort 从监听地址中解析纯端口号
func parsePort(addr string) string {
	if addr == "" {
		return "8080"
	}
	if strings.HasPrefix(addr, ":") {
		return addr[1:]
	}
	if parts := strings.Split(addr, ":"); len(parts) >= 2 {
		return parts[len(parts)-1]
	}
	return addr
}

// getLocalIP 获取本机局域网 IPv4 地址
func getLocalIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return ""
	}
	for _, address := range addrs {
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}
		}
	}
	return ""
}

// printServerURLs 打印本地与局域网访问地址
func printServerURLs(addr string) {
	port := parsePort(addr)
	fmt.Printf(" >> 本地访问:    http://127.0.0.1:%s (或 http://localhost:%s)\n", port, port)
	if ip := getLocalIP(); ip != "" && ip != "127.0.0.1" {
		fmt.Printf(" >> 局域网访问:  http://%s:%s\n", ip, port)
	}
}

// printOpsCommands 输出完整运维常用指令集 (包括启动、停止、重启、状态、日志)
func printOpsCommands(logFile string) {
	cmd := getExecCommand()
	fmt.Println("----------------------------------------------------------")
	fmt.Printf(" >> 运维常用指令:\n")
	fmt.Printf("    - 启动服务:    %s start\n", cmd)
	fmt.Printf("    - 查看状态:    %s status\n", cmd)
	fmt.Printf("    - 重启服务:    %s restart\n", cmd)
	fmt.Printf("    - 停止服务:    %s stop\n", cmd)
	if logFile != "" {
		fmt.Printf("    - 实时查看日志: tail -f %s\n", logFile)
	}
	fmt.Println("==========================================================")
}


// readPID 读取 PID 文件中的数值
func (m *Manager) readPID() (int, bool) {
	bytes, err := os.ReadFile(m.cfg.PIDFile)
	if err != nil {
		return 0, false
	}
	s := strings.TrimSpace(string(bytes))
	pid, err := strconv.Atoi(s)
	if err != nil || pid <= 0 {
		return 0, false
	}
	return pid, true
}

// SafeRemovePID 清除 PID 文件助手
func (m *Manager) SafeRemovePID() {
	_ = os.Remove(m.cfg.PIDFile)
}

// SuppressOutput 抑制标准输出（辅助函数）
func SuppressOutput() {
	os.Stdout, _ = os.Open(os.DevNull)
	os.Stderr, _ = os.Open(os.DevNull)
}

// EnsureLogWriter 获取日志输出写入器
func EnsureLogWriter(logPath string) (io.Writer, error) {
	_ = os.MkdirAll(filepath.Dir(logPath), 0755)
	return os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
}

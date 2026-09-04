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
		m.Status()
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
		fmt.Printf(">> [WARN] Godeniter 服务已经在后台运行中！(PID: %d)\n", pid)
		fmt.Printf(">> [TIP] 查看状态: go run main.go status (或 ./app status)\n")
		fmt.Printf(">> [TIP] 平滑重启: go run main.go restart (或 ./app restart)\n")
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

	// 智能推测二进制名称
	binName := filepath.Base(os.Args[0])
	if strings.Contains(os.Args[0], "go-build") || binName == "main" {
		binName = "./app"
	} else if !strings.HasPrefix(binName, "./") && !strings.HasPrefix(binName, "/") {
		binName = "./" + binName
	}

	fmt.Println("==========================================================")
	fmt.Printf(" >> Godeniter 2.0 服务已成功在后台启动 (Daemon Mode)!\n")
	fmt.Printf(" >> 进程 PID:    %d (已写入 %s)\n", pid, m.cfg.PIDFile)
	fmt.Printf(" >> 监听端口:    %s\n", addr)
	fmt.Printf(" >> 输出日志:    %s\n", m.cfg.LogFile)
	fmt.Println("----------------------------------------------------------")
	fmt.Printf(" >> 运维常用指令 (源码运行 / 二进制运行):\n")
	fmt.Printf("    - 查看状态: go run main.go status  (或 %s status)\n", binName)
	fmt.Printf("    - 停止服务: go run main.go stop    (或 %s stop)\n", binName)
	fmt.Printf("    - 重启服务: go run main.go restart (或 %s restart)\n", binName)
	fmt.Printf("    - 实时日志: tail -f %s\n", m.cfg.LogFile)
	fmt.Println("==========================================================")

	return nil

}

// Stop 优雅停止后台服务
func (m *Manager) Stop() error {
	pid, ok := m.readPID()
	if !ok {
		fmt.Printf(">> [STATUS] 服务未运行 (未检测到 PID 文件: %s)\n", m.cfg.PIDFile)
		return nil
	}

	if !checkProcess(pid) {
		_ = os.Remove(m.cfg.PIDFile)
		fmt.Printf(">> [STATUS] 目标进程 (PID: %d) 已不存在，已自动清理失效的 PID 文件。\n", pid)
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
			return nil
		}
	}

	// 若超时仍未退出，尝试强制清理
	_ = forceKillProcess(pid)
	_ = os.Remove(m.cfg.PIDFile)
	fmt.Printf(">> [WARN] 服务在超时时间内未自行退出，已执行终结。\n")
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
func (m *Manager) Status() {
	pid, ok := m.readPID()
	if !ok {
		fmt.Printf(">> [STATUS] Godeniter 服务状态: [未运行] (PID 文件不存在)\n")
		return
	}

	if !checkProcess(pid) {
		_ = os.Remove(m.cfg.PIDFile)
		fmt.Printf(">> [STATUS] Godeniter 服务状态: [已停止] (残留 PID 文件已自动清理)\n")
		return
	}

	fmt.Printf("==========================================================\n")
	fmt.Printf(" >> Godeniter 服务状态: [运行中 🟢]\n")
	fmt.Printf(" >> 运行 PID:    %d\n", pid)
	fmt.Printf(" >> PID 文件:    %s\n", m.cfg.PIDFile)
	fmt.Printf(" >> 日志文件:    %s\n", m.cfg.LogFile)
	fmt.Printf("==========================================================\n")
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

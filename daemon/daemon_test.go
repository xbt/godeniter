package daemon

import (
	"os"
	"strconv"
	"testing"
)

type mockRunner struct {
	called bool
	addr   string
}

func (m *mockRunner) Run(addr ...string) error {
	m.called = true
	if len(addr) > 0 {
		m.addr = addr[0]
	}
	return nil
}


func TestConfigDefault(t *testing.T) {
	cfg := DefaultConfig()
	if cfg.PIDFile != "./app.pid" {
		t.Errorf("默认 PIDFile 应为 ./app.pid，实际为: %s", cfg.PIDFile)
	}
	if cfg.LogFile != "./app.log" {
		t.Errorf("默认 LogFile 应为 ./app.log，实际为: %s", cfg.LogFile)
	}
	if cfg.Daemon {
		t.Errorf("默认 Daemon 状态应为 false")
	}
}

func TestManagerPIDLifecycle(t *testing.T) {
	tmpPID := "./test_daemon.pid"
	defer os.Remove(tmpPID)

	mgr := New(Config{
		PIDFile: tmpPID,
		LogFile: "./test_daemon.log",
	})
	defer os.Remove("./test_daemon.log")

	// 1. 测试未写入时读取
	if _, ok := mgr.readPID(); ok {
		t.Errorf("未写入 PID 时预期读取失败")
	}

	// 2. 模拟写入当前进程自身 PID
	currentPID := os.Getpid()
	_ = os.WriteFile(tmpPID, []byte(strconv.Itoa(currentPID)), 0644)

	pid, ok := mgr.readPID()
	if !ok || pid != currentPID {
		t.Errorf("读取 PID 失败，预期: %d, 实际: %d", currentPID, pid)
	}

	// 3. 验证当前进程存活检查
	if !checkProcess(currentPID) {
		t.Errorf("当前进程预期检查为存活状态")
	}

	// 4. 清理 PID 文件
	mgr.SafeRemovePID()
	if _, ok := mgr.readPID(); ok {
		t.Errorf("清理后 PID 文件应不存在")
	}
}

func TestRunWorker(t *testing.T) {
	tmpPID := "./test_worker.pid"
	defer os.Remove(tmpPID)

	mgr := New(Config{
		PIDFile: tmpPID,
		LogFile: "./test_worker.log",
	})
	defer os.Remove("./test_worker.log")

	mock := &mockRunner{}
	err := mgr.runWorker(mock, ":9090")
	if err != nil {
		t.Fatalf("runWorker 预期执行无错误: %v", err)
	}
	if !mock.called || mock.addr != ":9090" {
		t.Errorf("mockRunner 预期被调用且地址为 :9090")
	}

	// 验证 runWorker 执行完毕后 PID 文件自动清理
	if _, ok := mgr.readPID(); ok {
		t.Errorf("runWorker defer 预期清理 PID 文件")
	}
}

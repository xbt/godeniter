// Copyright (c) 2026 xbt. All rights reserved.
// Godeniter is licensed under the GNU General Public License v3.0 (GPL-3.0).
// See LICENSE file in the project root for full license information.

package cron

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"
)

// TestCronParser 测试自适应 5 段式与 6 段秒级解析
func TestCronParser(t *testing.T) {
	// 1. 经典 5 段式
	s5, err := Parse("0 3 * * *")
	if err != nil {
		t.Fatalf("解析 5 段式失败: %v", err)
	}
	base := time.Date(2026, 9, 9, 2, 0, 0, 0, time.Local)
	next5 := s5.Next(base)
	expected5 := time.Date(2026, 9, 9, 3, 0, 0, 0, time.Local)
	if !next5.Equal(expected5) {
		t.Errorf("5 段式时间推算不匹配: 期望 %v, 实际 %v", expected5, next5)
	}

	// 2. 现代 6 段秒级式
	s6, err := Parse("*/10 * * * * *")
	if err != nil {
		t.Fatalf("解析 6 段式失败: %v", err)
	}
	baseSec := time.Date(2026, 9, 9, 12, 0, 3, 0, time.Local)
	next6 := s6.Next(baseSec)
	expected6 := time.Date(2026, 9, 9, 12, 0, 10, 0, time.Local)
	if !next6.Equal(expected6) {
		t.Errorf("6 段秒级时间推算不匹配: 期望 %v, 实际 %v", expected6, next6)
	}

	// 3. @every 宏
	se, err := Parse("@every 5s")
	if err != nil {
		t.Fatalf("解析 @every 宏失败: %v", err)
	}
	nextE := se.Next(baseSec)
	if nextE.Sub(baseSec) != 5*time.Second {
		t.Errorf("@every 5s 时间间隔不匹配")
	}

	// 4. 自然语言释义
	if s5.HumanString() != "每天 03:00" {
		t.Errorf("自然语言解析异常: %s", s5.HumanString())
	}
	if s6.HumanString() != "每 10 秒" {
		t.Errorf("自然语言秒级解析异常: %s", s6.HumanString())
	}
}

// TestCronExecutionAndTrigger 测试真实运行、手动立即触发、暂停与恢复
func TestCronExecutionAndTrigger(t *testing.T) {
	c := New()

	var counter int32
	_, err := c.Add("test_counter", "测试递增任务", "* * * * * *", func() error {
		atomic.AddInt32(&counter, 1)
		return nil
	})
	if err != nil {
		t.Fatalf("添加任务失败: %v", err)
	}

	// 验证快照
	jobs := c.Jobs()
	if len(jobs) != 1 || jobs[0].ID != "test_counter" {
		t.Fatalf("任务快照不匹配: %+v", jobs)
	}

	// 手动立即触发
	err = c.Trigger("test_counter")
	if err != nil {
		t.Fatalf("手动触发失败: %v", err)
	}

	time.Sleep(50 * time.Millisecond)
	if atomic.LoadInt32(&counter) < 1 {
		t.Errorf("手动触发后任务未执行")
	}

	snap := c.GetJob("test_counter")
	if snap.LastStatus != "SUCCESS" || snap.RunCount < 1 {
		t.Errorf("任务快照未记录成功状态: %+v", snap)
	}

	// 测试暂停
	_ = c.Pause("test_counter")
	snapPaused := c.GetJob("test_counter")
	if snapPaused.Enabled {
		t.Errorf("任务暂停状态未生效")
	}

	// 测试恢复
	_ = c.Resume("test_counter")
	snapResumed := c.GetJob("test_counter")
	if !snapResumed.Enabled {
		t.Errorf("任务恢复状态未生效")
	}
}

// TestCronPanicRecovery 测试任务 Panic 安全隔离保护
func TestCronPanicRecovery(t *testing.T) {
	c := New()

	_, err := c.Add("panic_job", "异常任务", "* * * * * *", func() error {
		panic("模拟内存空指针异常")
	})
	if err != nil {
		t.Fatalf("添加异常任务失败: %v", err)
	}

	// 立即触发
	_ = c.Trigger("panic_job")
	time.Sleep(50 * time.Millisecond)

	snap := c.GetJob("panic_job")
	if snap.LastStatus != "FAILED" {
		t.Errorf("Panic 任务应标记为 FAILED，实际: %s", snap.LastStatus)
	}
	if snap.LastError == "" {
		t.Errorf("未捕获到 Panic 错误信息")
	}
}

// TestCronSchedulerLifecycle 测试调度循环与优雅退出
func TestCronSchedulerLifecycle(t *testing.T) {
	c := New()

	var tickCount int32
	_, _ = c.Every("every_sec", "每秒任务", 1*time.Second, func() error {
		atomic.AddInt32(&tickCount, 1)
		return nil
	})

	c.Start()
	if !c.IsRunning() {
		t.Fatalf("调度器启动后应处于运行态")
	}

	// 运行约 1.5 秒
	time.Sleep(1500 * time.Millisecond)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err := c.Stop(ctx)
	if err != nil {
		t.Fatalf("调度器停止失败: %v", err)
	}

	if c.IsRunning() {
		t.Errorf("调度器已停止但状态仍显示为运行中")
	}

	if atomic.LoadInt32(&tickCount) < 1 {
		t.Errorf("每秒调度未触发执行: count = %d", tickCount)
	}
}

// TestCronErrorHandling 测试普通 error 返回
func TestCronErrorHandling(t *testing.T) {
	c := New()

	_, _ = c.Add("err_job", "错误返回任务", "* * * * * *", func() error {
		return errors.New("网络连接超时")
	})

	_ = c.Trigger("err_job")
	time.Sleep(50 * time.Millisecond)

	snap := c.GetJob("err_job")
	if snap.LastStatus != "FAILED" || snap.LastError != "网络连接超时" {
		t.Errorf("任务错误状态未正确记录: %+v", snap)
	}
}

// TestCronUpdateJob 测试动态更新任务规格与名称
func TestCronUpdateJob(t *testing.T) {
	c := New()

	_, err := c.Add("backup", "旧备份任务", "0 0 3 * * *", func() error {
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}

	snap1 := c.GetJob("backup")
	if snap1.Spec != "0 0 3 * * *" || snap1.Name != "旧备份任务" {
		t.Fatalf("初始任务状态不符: %+v", snap1)
	}

	// 1. 合法更新为每天凌晨 4 点
	err = c.Update("backup", "新备份任务", "0 0 4 * * *")
	if err != nil {
		t.Fatalf("更新任务失败: %v", err)
	}

	snap2 := c.GetJob("backup")
	if snap2.Spec != "0 0 4 * * *" || snap2.Name != "新备份任务" {
		t.Fatalf("更新后任务规格未生效: %+v", snap2)
	}

	// 2. 非法语法更新应被拦截
	errBad := c.Update("backup", "非法表达式任务", "99 99 * * *")
	if errBad == nil {
		t.Fatalf("非法表达式预期报错，实际返回 nil")
	}

	// 3. 不存在的任务报错
	errNotExist := c.Update("not_exist", "不存在", "0 0 1 * * *")
	if errNotExist == nil {
		t.Fatalf("不存在的任务预期报错，实际返回 nil")
	}
}


// Copyright (c) 2026 xbt. All rights reserved.
// Godeniter is licensed under the GNU General Public License v3.0 (GPL-3.0).
// See LICENSE file in the project root for full license information.

package cron

import (
	"context"
	"fmt"
	"log"
	"runtime/debug"
	"sync"
	"time"
)

// JobFunc 定义任务执行函数的签名。支持返回 error 以便引擎记录日志与状态。
type JobFunc func() error

// Job 计划任务运行时实体。
type Job struct {
	ID          string        // 任务唯一标识 (如 "auto_backup")
	Name        string        // 任务展示名称
	Spec        string        // 原始表达式 (如 "*/30 * * * * *")
	HumanSpec   string        // 自然语言释义 (如 "每 30 秒")
	Schedule    Schedule      // 调度计算接口
	Fn          JobFunc       // 实际业务执行体
	Enabled     bool          // 启用/暂停状态
	PrevTime    time.Time     // 上次运行时间
	NextTime    time.Time     // 下次预定执行时间
	Duration    time.Duration // 上次执行耗时
	LastStatus  string        // SUCCESS / FAILED / RUNNING / IDLE
	LastError   string        // 最后一次执行的错误日志 (若有)
	RunCount    int64         // 累计调度次数
	isExecuting bool          // 防重入标识
	mu          sync.Mutex    // 实体互斥锁
}

// JobSnapshot 任务的只读状态快照，用于对外提供 JSON 序列化或视图呈现。
type JobSnapshot struct {
	ID          string        `json:"id"`
	Name        string        `json:"name"`
	Spec        string        `json:"spec"`
	HumanSpec   string        `json:"human_spec"`
	Enabled     bool          `json:"enabled"`
	PrevTime    string        `json:"prev_time"`
	NextTime    string        `json:"next_time"`
	DurationMs  float64       `json:"duration_ms"`
	DurationStr string        `json:"duration_str"`
	LastStatus  string        `json:"last_status"`
	LastError   string        `json:"last_error"`
	RunCount    int64         `json:"run_count"`
	IsRunning   bool          `json:"is_running"`
}

// Snapshot 获取当前任务状态的深拷贝快照。
func (j *Job) Snapshot() *JobSnapshot {
	j.mu.Lock()
	defer j.mu.Unlock()

	prevStr := "-"
	if !j.PrevTime.IsZero() {
		prevStr = j.PrevTime.Format("2006-01-02 15:04:05")
	}

	nextStr := "-"
	if !j.NextTime.IsZero() {
		nextStr = j.NextTime.Format("2006-01-02 15:04:05")
	}

	durMs := float64(j.Duration.Microseconds()) / 1000.0
	durStr := "-"
	if j.Duration > 0 {
		if j.Duration < time.Millisecond {
			durStr = fmt.Sprintf("%d µs", j.Duration.Microseconds())
		} else if j.Duration < time.Second {
			durStr = fmt.Sprintf("%.2f ms", durMs)
		} else {
			durStr = fmt.Sprintf("%.2f s", j.Duration.Seconds())
		}
	}

	return &JobSnapshot{
		ID:          j.ID,
		Name:        j.Name,
		Spec:        j.Spec,
		HumanSpec:   j.HumanSpec,
		Enabled:     j.Enabled,
		PrevTime:    prevStr,
		NextTime:    nextStr,
		DurationMs:  durMs,
		DurationStr: durStr,
		LastStatus:  j.LastStatus,
		LastError:   j.LastError,
		RunCount:    j.RunCount,
		IsRunning:   j.isExecuting,
	}
}

// Cron 纯 Go 秒级/分级全自适应计划任务调度引擎。
type Cron struct {
	jobs    []*Job
	jobMap  map[string]*Job
	running bool
	stop    chan struct{}
	wg      sync.WaitGroup
	mu      sync.RWMutex
}

// New 创建一个新的 Cron 调度器实例。
func New() *Cron {
	return &Cron{
		jobs:   make([]*Job, 0),
		jobMap: make(map[string]*Job),
		stop:   make(chan struct{}),
	}
}

// Add 注册一个标准 Cron 表达式任务。
// spec 支持 6 段秒级 (如 "*/10 * * * * *")、5 段分钟级 (如 "0 3 * * *") 或 @every 5s。
func (c *Cron) Add(id, name, spec string, fn JobFunc) (*Job, error) {
	schedule, err := Parse(spec)
	if err != nil {
		return nil, err
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	if _, exists := c.jobMap[id]; exists {
		return nil, fmt.Errorf("cron: 任务 ID [%s] 已存在", id)
	}

	now := time.Now().Truncate(time.Second)
	job := &Job{
		ID:         id,
		Name:       name,
		Spec:       spec,
		HumanSpec:  schedule.HumanString(),
		Schedule:   schedule,
		Fn:         fn,
		Enabled:    true,
		NextTime:   schedule.Next(now),
		LastStatus: "IDLE",
	}

	c.jobs = append(c.jobs, job)
	c.jobMap[id] = job
	return job, nil
}

// Every 注册一个基于时间间隔的快捷任务。
func (c *Cron) Every(id, name string, d time.Duration, fn JobFunc) (*Job, error) {
	sched := Every(d)

	c.mu.Lock()
	defer c.mu.Unlock()

	if _, exists := c.jobMap[id]; exists {
		return nil, fmt.Errorf("cron: 任务 ID [%s] 已存在", id)
	}

	now := time.Now().Truncate(time.Second)
	job := &Job{
		ID:         id,
		Name:       name,
		Spec:       sched.Expression(),
		HumanSpec:  sched.HumanString(),
		Schedule:   sched,
		Fn:         fn,
		Enabled:    true,
		NextTime:   sched.Next(now),
		LastStatus: "IDLE",
	}

	c.jobs = append(c.jobs, job)
	c.jobMap[id] = job
	return job, nil
}

// Remove 移除指定 ID 的任务。
func (c *Cron) Remove(id string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.jobMap, id)
	newJobs := make([]*Job, 0, len(c.jobs))
	for _, j := range c.jobs {
		if j.ID != id {
			newJobs = append(newJobs, j)
		}
	}
	c.jobs = newJobs
}

// Pause 暂停指定任务的自动调度。
func (c *Cron) Pause(id string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	job, exists := c.jobMap[id]
	if !exists {
		return fmt.Errorf("cron: 任务 [%s] 不存在", id)
	}

	job.mu.Lock()
	job.Enabled = false
	job.mu.Unlock()
	return nil
}

// Resume 恢复指定任务的自动调度。
func (c *Cron) Resume(id string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	job, exists := c.jobMap[id]
	if !exists {
		return fmt.Errorf("cron: 任务 [%s] 不存在", id)
	}

	job.mu.Lock()
	job.Enabled = true
	job.NextTime = job.Schedule.Next(time.Now())
	job.mu.Unlock()
	return nil
}

// Update 更新已注册任务的调度规格 (Cron 表达式) 与任务名称。
// 立即通过 Parse 校验新表达式，若合法则更新任务规格、自然语言释义，并根据新规则重算下次执行时间（秒级热生效）。
func (c *Cron) Update(id, newName, newSpec string) error {
	schedule, err := Parse(newSpec)
	if err != nil {
		return err
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	job, exists := c.jobMap[id]
	if !exists {
		return fmt.Errorf("cron: 任务 [%s] 不存在", id)
	}

	now := time.Now().Truncate(time.Second)
	job.mu.Lock()
	if newName != "" {
		job.Name = newName
	}
	job.Spec = newSpec
	job.HumanSpec = schedule.HumanString()
	job.Schedule = schedule
	if job.Enabled {
		job.NextTime = schedule.Next(now)
	}
	job.mu.Unlock()
	return nil
}

// Trigger 手动立即异步触发一次指定任务（执行耗时与状态即时更新，不影响其原定下次计划时间）。
func (c *Cron) Trigger(id string) error {
	c.mu.RLock()
	job, exists := c.jobMap[id]
	c.mu.RUnlock()

	if !exists {
		return fmt.Errorf("cron: 任务 [%s] 不存在", id)
	}

	go c.runJob(job, false)
	return nil
}

// Jobs 返回所有已注册任务的状态快照列表。
func (c *Cron) Jobs() []*JobSnapshot {
	c.mu.RLock()
	defer c.mu.RUnlock()

	snapshots := make([]*JobSnapshot, 0, len(c.jobs))
	for _, j := range c.jobs {
		snapshots = append(snapshots, j.Snapshot())
	}
	return snapshots
}

// GetJob 获取单个任务的状态快照。
func (c *Cron) GetJob(id string) *JobSnapshot {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if j, exists := c.jobMap[id]; exists {
		return j.Snapshot()
	}
	return nil
}

// Start 启动 Cron 调度器后台主循环 (默认 1 秒精度扫描)。
func (c *Cron) Start() {
	c.mu.Lock()
	if c.running {
		c.mu.Unlock()
		return
	}
	c.running = true
	c.stop = make(chan struct{})
	c.mu.Unlock()

	c.wg.Add(1)
	go c.runLoop()
}

// Stop 优雅关闭调度器并等待正在执行的任务安全完成。
func (c *Cron) Stop(ctx context.Context) error {
	c.mu.Lock()
	if !c.running {
		c.mu.Unlock()
		return nil
	}
	c.running = false
	close(c.stop)
	c.mu.Unlock()

	done := make(chan struct{})
	go func() {
		c.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// IsRunning 检查调度引擎是否处于运行状态。
func (c *Cron) IsRunning() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.running
}

func (c *Cron) runLoop() {
	defer c.wg.Done()

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case now := <-ticker.C:
			c.checkAndRun(now)
		case <-c.stop:
			return
		}
	}
}

func (c *Cron) checkAndRun(now time.Time) {
	c.mu.RLock()
	jobs := make([]*Job, len(c.jobs))
	copy(jobs, c.jobs)
	c.mu.RUnlock()

	nowSec := now.Truncate(time.Second)

	for _, job := range jobs {
		job.mu.Lock()
		if !job.Enabled {
			job.mu.Unlock()
			continue
		}

		if !job.NextTime.IsZero() && (nowSec.After(job.NextTime) || nowSec.Equal(job.NextTime)) {
			// 更新下一次触发时间
			job.NextTime = job.Schedule.Next(nowSec)
			job.mu.Unlock()

			// 异步派发执行
			go c.runJob(job, true)
		} else {
			job.mu.Unlock()
		}
	}
}

func (c *Cron) runJob(job *Job, updateNext bool) {
	job.mu.Lock()
	if job.isExecuting {
		job.mu.Unlock()
		return // 任务尚未完成，防重入跳过
	}
	job.isExecuting = true
	job.LastStatus = "RUNNING"
	job.PrevTime = time.Now()
	job.mu.Unlock()

	c.wg.Add(1)
	defer func() {
		c.wg.Done()
	}()

	start := time.Now()
	var runErr error

	// 捕获任务 Panic，保证主程序与调度引擎绝对不崩溃
	func() {
		defer func() {
			if r := recover(); r != nil {
				runErr = fmt.Errorf("panic: %v\nstack:\n%s", r, string(debug.Stack()))
				log.Printf("[CRON ERROR] 任务 [%s - %s] 执行中发生 Panic: %v\n", job.ID, job.Name, r)
			}
		}()

		if job.Fn != nil {
			runErr = job.Fn()
		}
	}()

	elapsed := time.Since(start)

	job.mu.Lock()
	job.Duration = elapsed
	job.RunCount++
	job.isExecuting = false

	if runErr != nil {
		job.LastStatus = "FAILED"
		job.LastError = runErr.Error()
	} else {
		job.LastStatus = "SUCCESS"
		job.LastError = ""
	}
	job.mu.Unlock()
}

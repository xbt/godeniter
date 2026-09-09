// Copyright (c) 2026 xbt. All rights reserved.
// Godeniter is licensed under the GNU General Public License v3.0 (GPL-3.0).
// See LICENSE file in the project root for full license information.

package cron

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// Schedule 定义了计算下一次任务执行时间的标准接口。
type Schedule interface {
	Next(time.Time) time.Time
	Expression() string
	HumanString() string
}

// EverySchedule 纯时间间隔调度器 (如 @every 5s, @every 1m)
type EverySchedule struct {
	duration time.Duration
}

// Every 创建一个基于时间间隔的调度计划。
func Every(d time.Duration) Schedule {
	if d < time.Second {
		d = time.Second // 最小粒度为 1 秒
	}
	return &EverySchedule{duration: d}
}

func (s *EverySchedule) Next(t time.Time) time.Time {
	return t.Truncate(time.Second).Add(s.duration)
}

func (s *EverySchedule) Expression() string {
	return fmt.Sprintf("@every %s", s.duration.String())
}

func (s *EverySchedule) HumanString() string {
	d := s.duration
	if d < time.Minute {
		return fmt.Sprintf("每 %d 秒", int(d.Seconds()))
	}
	if d < time.Hour {
		return fmt.Sprintf("每 %d 分钟", int(d.Minutes()))
	}
	return fmt.Sprintf("每 %d 小时", int(d.Hours()))
}

// SpecSchedule 标准 Cron 表达式调度器 (支持 5 段分钟级与 6 段秒级)
type SpecSchedule struct {
	expr     string
	humanStr string
	seconds  uint64 // 位图: 0-59 (第 0 位代表 0 秒)
	minutes  uint64 // 位图: 0-59
	hours    uint32 // 位图: 0-23
	days     uint32 // 位图: 1-31 (第 1 位代表 1 号)
	months   uint16 // 位图: 1-12
	weekdays uint8  // 位图: 0-6 (0=周日, 6=周六)
}

// Parse 解析 Cron 表达式字符串，自适应 6 段秒级、5 段 Linux 格式与常用宏。
func Parse(spec string) (Schedule, error) {
	spec = strings.TrimSpace(spec)
	if spec == "" {
		return nil, fmt.Errorf("cron: 表达式不能为空")
	}

	// 1. 处理 @ 符号引导的语义快捷宏
	if strings.HasPrefix(spec, "@") {
		lower := strings.ToLower(spec)
		if strings.HasPrefix(lower, "@every ") {
			dStr := strings.TrimSpace(spec[7:])
			d, err := time.ParseDuration(dStr)
			if err != nil {
				return nil, fmt.Errorf("cron: 无效的 @every 时间间隔 [%s]: %w", dStr, err)
			}
			return Every(d), nil
		}

		switch lower {
		case "@yearly", "@annually":
			return parseSpecParts("0 0 0 1 1 *", spec, "每年 1月1日 00:00")
		case "@monthly":
			return parseSpecParts("0 0 0 1 * *", spec, "每月 1日 00:00")
		case "@weekly":
			return parseSpecParts("0 0 0 * * 0", spec, "每周日 00:00")
		case "@daily", "@midnight":
			return parseSpecParts("0 0 0 * * *", spec, "每天 00:00:00")
		case "@hourly":
			return parseSpecParts("0 0 * * * *", spec, "每小时整点")
		default:
			return nil, fmt.Errorf("cron: 未知的描述宏: %s", spec)
		}
	}

	fields := strings.Fields(spec)
	var standardExpr string
	var humanStr string

	if len(fields) == 5 {
		// 经典 Linux 5 段式: 分 时 日 月 周 -> 补齐第 0 秒为 0
		standardExpr = "0 " + spec
		humanStr = describeCron5(fields)
	} else if len(fields) == 6 {
		// 现代秒级 6 段式: 秒 分 时 日 月 周
		standardExpr = spec
		humanStr = describeCron6(fields)
	} else {
		return nil, fmt.Errorf("cron: 表达式必须为 5 段(经典分级)或 6 段(秒级)，当前包含 %d 个字段: %s", len(fields), spec)
	}

	return parseSpecParts(standardExpr, spec, humanStr)
}

func parseSpecParts(sixPartExpr, originalExpr, humanStr string) (Schedule, error) {
	fields := strings.Fields(sixPartExpr)
	if len(fields) != 6 {
		return nil, fmt.Errorf("cron: 内部格式错误: %s", sixPartExpr)
	}

	sec, err := parseField(fields[0], 0, 59, nil)
	if err != nil {
		return nil, fmt.Errorf("秒字段错误: %w", err)
	}
	min, err := parseField(fields[1], 0, 59, nil)
	if err != nil {
		return nil, fmt.Errorf("分字段错误: %w", err)
	}
	hr, err := parseField(fields[2], 0, 23, nil)
	if err != nil {
		return nil, fmt.Errorf("时字段错误: %w", err)
	}
	dom, err := parseField(fields[3], 1, 31, nil)
	if err != nil {
		return nil, fmt.Errorf("日字段错误: %w", err)
	}
	mon, err := parseField(fields[4], 1, 12, monthNames)
	if err != nil {
		return nil, fmt.Errorf("月字段错误: %w", err)
	}
	dow, err := parseField(fields[5], 0, 6, dayNames)
	if err != nil {
		return nil, fmt.Errorf("周字段错误: %w", err)
	}

	return &SpecSchedule{
		expr:     originalExpr,
		humanStr: humanStr,
		seconds:  sec,
		minutes:  min,
		hours:    uint32(hr),
		days:     uint32(dom),
		months:   uint16(mon),
		weekdays: uint8(dow),
	}, nil
}

func (s *SpecSchedule) Expression() string {
	return s.expr
}

func (s *SpecSchedule) HumanString() string {
	if s.humanStr != "" {
		return s.humanStr
	}
	return s.expr
}

// Next 按照给定时间寻找下一个匹配的时间点 (精度精确到秒)
func (s *SpecSchedule) Next(t time.Time) time.Time {
	next := t.Truncate(time.Second).Add(1 * time.Second)
	maxYear := next.Year() + 5

WRAP:
	if next.Year() > maxYear {
		return time.Time{}
	}

	// 1. 匹配月份 (1-12)
	for 1<<uint(next.Month())&s.months == 0 {
		next = time.Date(next.Year(), next.Month()+1, 1, 0, 0, 0, 0, next.Location())
		if next.Year() > maxYear {
			return time.Time{}
		}
	}

	// 2. 匹配日期
	for !s.dayMatches(next) {
		next = time.Date(next.Year(), next.Month(), next.Day()+1, 0, 0, 0, 0, next.Location())
		if next.Year() > maxYear {
			return time.Time{}
		}
		if 1<<uint(next.Month())&s.months == 0 {
			goto WRAP
		}
	}

	// 3. 匹配小时 (0-23)
	for 1<<uint(next.Hour())&s.hours == 0 {
		next = next.Truncate(time.Hour).Add(time.Hour)
		if next.Hour() == 0 {
			goto WRAP
		}
	}

	// 4. 匹配分钟 (0-59)
	for 1<<uint(next.Minute())&s.minutes == 0 {
		next = next.Truncate(time.Minute).Add(time.Minute)
		if next.Minute() == 0 {
			goto WRAP
		}
	}

	// 5. 匹配秒 (0-59)
	for 1<<uint(next.Second())&s.seconds == 0 {
		next = next.Add(time.Second)
		if next.Second() == 0 {
			goto WRAP
		}
	}

	return next
}

func (s *SpecSchedule) dayMatches(t time.Time) bool {
	domMatch := 1<<uint(t.Day())&s.days != 0
	dowMatch := 1<<uint(t.Weekday())&uint32(s.weekdays) != 0
	return domMatch && dowMatch
}

func parseField(field string, min, max int, names map[string]int) (uint64, error) {
	var bits uint64

	for _, expr := range strings.Split(field, ",") {
		expr = strings.TrimSpace(expr)
		if expr == "" {
			continue
		}

		var rangeAndStep = strings.Split(expr, "/")
		var low, high int
		var step = 1
		var err error

		if len(rangeAndStep) == 2 {
			step, err = strconv.Atoi(rangeAndStep[1])
			if err != nil || step <= 0 {
				return 0, fmt.Errorf("无效的步长: %s", rangeAndStep[1])
			}
		}

		rangeStr := rangeAndStep[0]
		if rangeStr == "*" || rangeStr == "?" {
			low = min
			high = max
		} else if strings.Contains(rangeStr, "-") {
			parts := strings.Split(rangeStr, "-")
			if len(parts) != 2 {
				return 0, fmt.Errorf("无效的范围: %s", rangeStr)
			}
			low, err = parseItem(parts[0], names)
			if err != nil {
				return 0, err
			}
			high, err = parseItem(parts[1], names)
			if err != nil {
				return 0, err
			}
		} else {
			low, err = parseItem(rangeStr, names)
			if err != nil {
				return 0, err
			}
			if len(rangeAndStep) == 1 {
				high = low
			} else {
				high = max
			}
		}

		if max == 6 && min == 0 {
			if low == 7 {
				low = 0
			}
			if high == 7 {
				high = 0
			}
		}

		if low < min || low > max || high < min || high > max || low > high {
			return 0, fmt.Errorf("值 [%d-%d] 超出有效范围 [%d-%d]", low, high, min, max)
		}

		for i := low; i <= high; i += step {
			bits |= 1 << uint(i)
		}
	}

	return bits, nil
}

func parseItem(val string, names map[string]int) (int, error) {
	if names != nil {
		if n, ok := names[strings.ToUpper(val)]; ok {
			return n, nil
		}
	}
	n, err := strconv.Atoi(val)
	if err != nil {
		return 0, fmt.Errorf("无效的数值: %s", val)
	}
	return n, nil
}

var monthNames = map[string]int{
	"JAN": 1, "FEB": 2, "MAR": 3, "APR": 4, "MAY": 5, "JUN": 6,
	"JUL": 7, "AUG": 8, "SEP": 9, "OCT": 10, "NOV": 11, "DEC": 12,
}

var dayNames = map[string]int{
	"SUN": 0, "MON": 1, "TUE": 2, "WED": 3, "THU": 4, "FRI": 5, "SAT": 6,
}

func describeCron5(f []string) string {
	if f[0] == "*" && f[1] == "*" && f[2] == "*" && f[3] == "*" && f[4] == "*" {
		return "每分钟"
	}
	if strings.HasPrefix(f[0], "*/") && f[1] == "*" {
		return fmt.Sprintf("每 %s 分钟", strings.TrimPrefix(f[0], "*/"))
	}
	if f[0] == "0" && f[1] == "*" && f[2] == "*" && f[3] == "*" && f[4] == "*" {
		return "每小时整点"
	}
	if f[2] == "*" && f[3] == "*" && f[4] == "*" {
		if _, err := strconv.Atoi(f[0]); err == nil {
			if _, err2 := strconv.Atoi(f[1]); err2 == nil {
				return fmt.Sprintf("每天 %02s:%02s", f[1], f[0])
			}
		}
	}
	return strings.Join(f, " ")
}

func describeCron6(f []string) string {
	if strings.HasPrefix(f[0], "*/") && f[1] == "*" && f[2] == "*" {
		return fmt.Sprintf("每 %s 秒", strings.TrimPrefix(f[0], "*/"))
	}
	if f[0] == "0" && f[1] == "0" && f[3] == "*" && f[4] == "*" && f[5] == "*" {
		if _, err := strconv.Atoi(f[2]); err == nil {
			return fmt.Sprintf("每天 %02s:00:00", f[2])
		}
	}
	return strings.Join(f, " ")
}

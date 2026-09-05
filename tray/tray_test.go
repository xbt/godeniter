// Copyright (c) 2026 xbt. All rights reserved.
// Godeniter is licensed under the GNU General Public License v3.0 (GPL-3.0).
// See LICENSE file in the project root for full license information.

package tray

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultOptions(t *testing.T) {
	opts := DefaultOptions()
	if opts.Title != "Godeniter" {
		t.Fatalf("期望默认 Title 为 'Godeniter'，得到: %s", opts.Title)
	}
	if opts.URL != "http://127.0.0.1:8080" {
		t.Fatalf("期望默认 URL 为 'http://127.0.0.1:8080'，得到: %s", opts.URL)
	}
	if opts.AppDir == "" {
		t.Fatalf("期望默认 AppDir 非空")
	}
}

func TestGetExecutableDir(t *testing.T) {
	dir := GetExecutableDir()
	if dir == "" {
		t.Fatalf("GetExecutableDir 返回空路径")
	}
	// 验证路径存在
	if fi, err := os.Stat(dir); err != nil || !fi.IsDir() {
		t.Fatalf("GetExecutableDir 返回的路径不是有效目录: %s", dir)
	}
}

func TestOpenURLValidation(t *testing.T) {
	err := OpenURL("")
	if err == nil {
		t.Fatalf("空 URL 应该返回错误")
	}
}

func TestOpenFolderValidation(t *testing.T) {
	dir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	abs, err := filepath.Abs(dir)
	if err != nil {
		t.Fatal(err)
	}
	if abs == "" {
		t.Fatalf("绝对路径为空")
	}
}

func TestBuildFullMenuItems(t *testing.T) {
	opts := Options{
		Title:   "TestApp",
		URL:     "http://127.0.0.1:9090",
		AppDir:  ".",
		Version: "v1.2.3",
		Menus: []MenuItem{
			{Title: "自定义菜单1", OnClick: func() {}},
			{Title: "自定义菜单2", OnClick: func() {}},
		},
	}

	items := buildFullMenuItems(opts)
	if len(items) < 5 {
		t.Fatalf("菜单项数量不足，实际: %d", len(items))
	}

	// 验证包含管理后台、应用目录、关于系统、退出
	var hasAdmin, hasDir, hasAbout, hasCustom, hasQuit bool
	for _, it := range items {
		if it.Title == "🌐 打开管理后台" {
			hasAdmin = true
		}
		if it.Title == "📁 打开应用目录" {
			hasDir = true
		}
		if it.Title == "ℹ️ 关于系统" {
			hasAbout = true
		}
		if it.Title == "自定义菜单1" {
			hasCustom = true
		}
		if it.Title == "⏹️ 退出程序" {
			hasQuit = true
		}
	}

	if !hasAdmin || !hasDir || !hasAbout || !hasCustom || !hasQuit {
		t.Fatalf("菜单项未完全组装成功: admin=%v, dir=%v, about=%v, custom=%v, quit=%v",
			hasAdmin, hasDir, hasAbout, hasCustom, hasQuit)
	}
}

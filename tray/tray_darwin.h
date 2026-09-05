// Copyright (c) 2026 xbt. All rights reserved.
// Godeniter is licensed under the GNU General Public License v3.0 (GPL-3.0).
// See LICENSE file in the project root for full license information.

#ifndef TRAY_DARWIN_H
#define TRAY_DARWIN_H

#include <stddef.h>

// MenuItem 跨语言结构体
typedef struct {
    const char* title;
    int is_separator;
    int disabled;
    int checked;
    int callback_id;
} TrayMenuItemC;

// 初始化 macOS NSApplication 并设置为 Accessory 模式 (不在 Dock 栏显示，仅在右上角状态栏显示)
void native_init_app(void);

// 创建顶部状态栏图标
void native_create_status_bar(const void* icon_bytes, size_t icon_len, const char* fallback_title, const char* tooltip);

// 构建与更新菜单项
void native_update_menu(TrayMenuItemC* items, int count);

// 启动 macOS 主事件循环
void native_run_loop(void);

// 退出 macOS 主事件循环
void native_quit_loop(void);

// 弹出原生 macOS 提示对话框 (关于系统等)
void native_show_alert(const char* title, const char* message);

#endif // TRAY_DARWIN_H

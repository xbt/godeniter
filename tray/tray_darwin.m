//go:build darwin

// Copyright (c) 2026 xbt. All rights reserved.
// Godeniter is licensed under the GNU General Public License v3.0 (GPL-3.0).
// See LICENSE file in the project root for full license information.

#import <Cocoa/Cocoa.h>
#import "tray_darwin.h"

extern void trayMenuItemCallback(int id);

@interface GodeniterTrayActionTarget : NSObject
- (void)menuItemClicked:(id)sender;
@end

@implementation GodeniterTrayActionTarget
- (void)menuItemClicked:(id)sender {
    NSMenuItem *item = (NSMenuItem *)sender;
    int cb_id = (int)[item tag];
    trayMenuItemCallback(cb_id);
}
@end

static NSStatusItem *globalStatusItem = nil;
static GodeniterTrayActionTarget *globalActionTarget = nil;

void native_init_app(void) {
    @autoreleasepool {
        [NSApplication sharedApplication];
        // 若在终端直接运行二进制 (非 .app Bundle)，保持 Accessory 模式
        CFBundleRef mainBundle = CFBundleGetMainBundle();
        CFDictionaryRef infoDict = mainBundle ? CFBundleGetInfoDictionary(mainBundle) : NULL;
        if (!infoDict) {
            [NSApp setActivationPolicy:NSApplicationActivationPolicyAccessory];
        }
        if (!globalActionTarget) {
            globalActionTarget = [[GodeniterTrayActionTarget alloc] init];
        }
        [NSApp finishLaunching];
    }
}

void native_create_status_bar(const void* icon_bytes, size_t icon_len, const char* fallback_title, const char* tooltip) {
    @autoreleasepool {
        if (!globalStatusItem) {
            // 使用自适应宽度 (纯图标模式下仅占约 22px，完美适配 MacBook 刘海屏)
            globalStatusItem = [[NSStatusBar systemStatusBar] statusItemWithLength:NSVariableStatusItemLength];
        }
        NSStatusBarButton *button = [globalStatusItem button];
        if (tooltip && strlen(tooltip) > 0) {
            [button setToolTip:[NSString stringWithUTF8String:tooltip]];
        }

        BOOL iconSet = NO;

        if (icon_bytes && icon_len > 0) {
            NSData *data = [NSData dataWithBytes:icon_bytes length:icon_len];
            NSImage *img = [[NSImage alloc] initWithData:data];
            if (img && [img isValid]) {
                [img setSize:NSMakeSize(18, 18)];
                [img setTemplate:YES];
                [button setImage:img];
                [button setImagePosition:NSImageOnly];
                [button setTitle:@""];
                iconSet = YES;
            }
        }

        if (!iconSet) {
            // 回退方案: 仅展示单个紧凑 Emoji 🚀 (约 18-20px 宽)，杜绝任何长标题被刘海屏截断隐藏
            [button setTitle:@"🚀"];
            [button setImage:nil];
        }
    }
}

void native_update_menu(TrayMenuItemC* items, int count) {
    @autoreleasepool {
        if (!globalStatusItem) return;
        NSMenu *menu = [[NSMenu alloc] init];
        [menu setAutoenablesItems:NO];
        for (int i = 0; i < count; i++) {
            if (items[i].is_separator) {
                [menu addItem:[NSMenuItem separatorItem]];
            } else {
                NSString *title = [NSString stringWithUTF8String:items[i].title];
                NSMenuItem *item = [[NSMenuItem alloc] initWithTitle:title action:@selector(menuItemClicked:) keyEquivalent:@""];
                [item setTarget:globalActionTarget];
                [item setTag:items[i].callback_id];
                if (items[i].disabled) {
                    [item setEnabled:NO];
                } else {
                    [item setEnabled:YES];
                }
                if (items[i].checked) {
                    [item setState:NSControlStateValueOn];
                } else {
                    [item setState:NSControlStateValueOff];
                }
                [menu addItem:item];
            }
        }
        [globalStatusItem setMenu:menu];
    }
}

void native_run_loop(void) {
    @autoreleasepool {
        [NSApp run];
    }
}

void native_quit_loop(void) {
    dispatch_async(dispatch_get_main_queue(), ^{
        if (globalStatusItem) {
            [[NSStatusBar systemStatusBar] removeStatusItem:globalStatusItem];
            globalStatusItem = nil;
        }
        [NSApp stop:nil];
        NSEvent* event = [NSEvent otherEventWithType:NSEventTypeApplicationDefined
                                            location:NSMakePoint(0, 0)
                                       modifierFlags:0
                                           timestamp:0
                                        windowNumber:0
                                             context:nil
                                             subtype:0
                                               data1:0
                                               data2:0];
        [NSApp postEvent:event atStart:YES];
    });
}

void native_show_alert(const char* title, const char* message) {
    void (^showAlertBlock)(void) = ^{
        @autoreleasepool {
            [NSApplication sharedApplication];
            NSAlert *alert = [[NSAlert alloc] init];
            [alert setMessageText:[NSString stringWithUTF8String:title]];
            [alert setInformativeText:[NSString stringWithUTF8String:message]];
            [alert setAlertStyle:NSAlertStyleCritical];
            [alert addButtonWithTitle:@"确定"];
            [[alert window] setLevel:NSFloatingWindowLevel];
            [alert runModal];
        }
    };

    if ([NSThread isMainThread]) {
        showAlertBlock();
    } else {
        dispatch_sync(dispatch_get_main_queue(), showAlertBlock);
    }
}

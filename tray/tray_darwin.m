//go:build darwin

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
        // 设置为 Accessory 模式: 状态栏常驻，不占用 Dock 栏图标
        [NSApp setActivationPolicy:NSApplicationActivationPolicyAccessory];
        if (!globalActionTarget) {
            globalActionTarget = [[GodeniterTrayActionTarget alloc] init];
        }
        [NSApp finishLaunching];
    }
}

void native_create_status_bar(const void* icon_bytes, size_t icon_len, const char* fallback_title, const char* tooltip) {
    @autoreleasepool {
        if (!globalStatusItem) {
            // 使用自适应宽度，避免在刘海屏或文字较多时被系统截断隐藏
            globalStatusItem = [[NSStatusBar systemStatusBar] statusItemWithLength:NSVariableStatusItemLength];
        }
        NSStatusBarButton *button = [globalStatusItem button];
        if (tooltip && strlen(tooltip) > 0) {
            [button setToolTip:[NSString stringWithUTF8String:tooltip]];
        }

        NSString *title = (fallback_title && strlen(fallback_title) > 0) ? [NSString stringWithUTF8String:fallback_title] : @"Godeniter";
        BOOL iconSet = NO;

        if (icon_bytes && icon_len > 0) {
            NSData *data = [NSData dataWithBytes:icon_bytes length:icon_len];
            NSImage *img = [[NSImage alloc] initWithData:data];
            if (img && [img isValid]) {
                [img setSize:NSMakeSize(18, 18)];
                // 保持原图色彩，不盲目 setTemplate:YES，避免彩色图标在 Dark Mode 深色菜单栏下变黑隐形
                [button setImage:img];
                [button setImagePosition:NSImageLeft];
                [button setTitle:[NSString stringWithFormat:@" %@", title]];
                iconSet = YES;
            }
        }

        if (!iconSet) {
            // 回退方案: 图标与标题结合，Emoji 🚀 无论浅色还是深色模式均 100% 显眼可见
            [button setTitle:[NSString stringWithFormat:@"🚀 %@", title]];
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
    dispatch_async(dispatch_get_main_queue(), ^{
        @autoreleasepool {
            NSAlert *alert = [[NSAlert alloc] init];
            [alert setMessageText:[NSString stringWithUTF8String:title]];
            [alert setInformativeText:[NSString stringWithUTF8String:message]];
            [alert setAlertStyle:NSAlertStyleInformational];
            [alert addButtonWithTitle:@"确定"];
            [[alert window] setLevel:NSFloatingWindowLevel];
            [alert runModal];
        }
    });
}

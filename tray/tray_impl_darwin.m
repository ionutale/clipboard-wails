#include <Cocoa/Cocoa.h>

// Go callbacks (declared in tray_darwin.go)
extern void goTrayShowWindow(void);
extern void goTrayItemClicked(int index);
extern void goTrayClearHistory(void);
extern void goTrayQuit(void);

@interface CMTrayDelegate : NSObject
@end

@implementation CMTrayDelegate

- (void)handleOpen:(id)sender {
    goTrayShowWindow();
}

- (void)handleItem:(NSMenuItem *)sender {
    goTrayItemClicked((int)sender.tag);
}

- (void)handleClear:(id)sender {
    goTrayClearHistory();
}

- (void)handleQuit:(id)sender {
    goTrayQuit();
}

@end

static CMTrayDelegate *g_delegate;
static NSStatusItem  *g_statusItem;
static NSMenu        *g_menu;

void cmCreateStatusItem(void) {
    dispatch_async(dispatch_get_main_queue(), ^{
        g_delegate   = [[CMTrayDelegate alloc] init];
        g_statusItem = [[NSStatusBar systemStatusBar] statusItemWithLength:NSVariableStatusItemLength];
        g_statusItem.button.title = @"📋";

        g_menu = [[NSMenu alloc] initWithTitle:@"Clipboard"];
        g_statusItem.menu = g_menu;

        // Seed the menu so it's valid before the first UpdateMenu call.
        NSMenuItem *open = [[NSMenuItem alloc] initWithTitle:@"Open Clipboard Manager"
                                                      action:@selector(handleOpen:)
                                               keyEquivalent:@""];
        open.target = g_delegate;
        [g_menu addItem:open];
        [g_menu addItem:[NSMenuItem separatorItem]];

        NSMenuItem *quit = [[NSMenuItem alloc] initWithTitle:@"Quit"
                                                      action:@selector(handleQuit:)
                                               keyEquivalent:@"q"];
        quit.target = g_delegate;
        [g_menu addItem:quit];
    });
}

// cmUpdateMenu rebuilds the dropdown with the supplied item titles.
// titles is a C array of null-terminated UTF-8 strings; count is its length.
// Ownership of titles[] stays with the caller — we copy here before dispatching.
void cmUpdateMenu(const char **titles, int count) {
    // Copy titles into an NSArray immediately (before the async block runs).
    NSMutableArray *labels = [NSMutableArray arrayWithCapacity:count];
    for (int i = 0; i < count; i++) {
        [labels addObject:[NSString stringWithUTF8String:titles[i]]];
    }

    dispatch_async(dispatch_get_main_queue(), ^{
        [g_menu removeAllItems];

        // "Open" item
        NSMenuItem *open = [[NSMenuItem alloc] initWithTitle:@"Open Clipboard Manager"
                                                      action:@selector(handleOpen:)
                                               keyEquivalent:@""];
        open.target = g_delegate;
        [g_menu addItem:open];

        if (labels.count > 0) {
            [g_menu addItem:[NSMenuItem separatorItem]];

            for (NSInteger i = 0; i < (NSInteger)labels.count; i++) {
                NSMenuItem *item = [[NSMenuItem alloc] initWithTitle:labels[i]
                                                              action:@selector(handleItem:)
                                                       keyEquivalent:@""];
                item.target = g_delegate;
                item.tag    = i;
                [g_menu addItem:item];
            }

            [g_menu addItem:[NSMenuItem separatorItem]];

            NSMenuItem *clear = [[NSMenuItem alloc] initWithTitle:@"Clear History"
                                                           action:@selector(handleClear:)
                                                    keyEquivalent:@""];
            clear.target = g_delegate;
            [g_menu addItem:clear];
        }

        [g_menu addItem:[NSMenuItem separatorItem]];

        NSMenuItem *quit = [[NSMenuItem alloc] initWithTitle:@"Quit"
                                                      action:@selector(handleQuit:)
                                               keyEquivalent:@"q"];
        quit.target = g_delegate;
        [g_menu addItem:quit];
    });
}

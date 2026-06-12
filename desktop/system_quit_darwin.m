//go:build darwin

#import <Cocoa/Cocoa.h>
#import <objc/runtime.h>

extern void DeepSeekOrcaMarkSystemQuit(void);

static NSApplicationTerminateReply (*originalApplicationShouldTerminate)(id, SEL, NSApplication *);
static void (*originalWailsContextQuit)(id, SEL);

static NSApplicationTerminateReply deepseekOrcaApplicationShouldTerminate(id self, SEL _cmd, NSApplication *sender) {
    DeepSeekOrcaMarkSystemQuit();
    if (originalApplicationShouldTerminate != NULL) {
        return originalApplicationShouldTerminate(self, _cmd, sender);
    }
    return NSTerminateNow;
}

static void deepseekOrcaWailsContextQuit(id self, SEL _cmd) {
    DeepSeekOrcaMarkSystemQuit();
    if (originalWailsContextQuit != NULL) {
        originalWailsContextQuit(self, _cmd);
    }
}

void installDeepSeekOrcaSystemQuitHook(void) {
    Class appDelegate = NSClassFromString(@"AppDelegate");
    SEL selector = @selector(applicationShouldTerminate:);
    Method method = appDelegate == Nil ? NULL : class_getInstanceMethod(appDelegate, selector);
    if (method != NULL) {
        IMP replacement = (IMP)deepseekOrcaApplicationShouldTerminate;
        IMP previous = method_setImplementation(method, replacement);
        originalApplicationShouldTerminate = (NSApplicationTerminateReply (*)(id, SEL, NSApplication *))previous;
    }

    Class wailsContext = NSClassFromString(@"WailsContext");
    SEL quitSelector = @selector(Quit);
    Method quitMethod = wailsContext == Nil ? NULL : class_getInstanceMethod(wailsContext, quitSelector);
    if (quitMethod != NULL) {
        IMP replacement = (IMP)deepseekOrcaWailsContextQuit;
        IMP previous = method_setImplementation(quitMethod, replacement);
        originalWailsContextQuit = (void (*)(id, SEL))previous;
    }
}

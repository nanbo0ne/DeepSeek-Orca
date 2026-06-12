//go:build darwin

package main

/*
#cgo darwin LDFLAGS: -framework Cocoa
void installDeepSeekOrcaSystemQuitHook(void);
*/
import "C"

import "sync"

var installSystemQuitHookOnce sync.Once

func installSystemQuitHook() {
	installSystemQuitHookOnce.Do(func() {
		C.installDeepSeekOrcaSystemQuitHook()
	})
}

//export DeepSeekOrcaMarkSystemQuit
func DeepSeekOrcaMarkSystemQuit() {
	markSystemQuitRequested()
}

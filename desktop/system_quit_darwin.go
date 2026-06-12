//go:build darwin

package main

/*
#cgo darwin LDFLAGS: -framework Cocoa
void installDeepSeek-OrcaSystemQuitHook(void);
*/
import "C"

import "sync"

var installSystemQuitHookOnce sync.Once

func installSystemQuitHook() {
	installSystemQuitHookOnce.Do(func() {
		C.installDeepSeek-OrcaSystemQuitHook()
	})
}

//export DeepSeek-OrcaMarkSystemQuit
func DeepSeek-OrcaMarkSystemQuit() {
	markSystemQuitRequested()
}

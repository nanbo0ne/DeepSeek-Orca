//go:build darwin

package main

/*
#cgo darwin LDFLAGS: -framework Cocoa
void installDeepCodeSystemQuitHook(void);
*/
import "C"

import "sync"

var installSystemQuitHookOnce sync.Once

func installSystemQuitHook() {
	installSystemQuitHookOnce.Do(func() {
		C.installDeepCodeSystemQuitHook()
	})
}

//export DeepCodeMarkSystemQuit
func DeepCodeMarkSystemQuit() {
	markSystemQuitRequested()
}

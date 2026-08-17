//go:build darwin

package main

/*
#cgo darwin LDFLAGS: -framework Cocoa
void installOrcaSystemQuitHook(void);
*/
import "C"

import "sync"

var installSystemQuitHookOnce sync.Once

func installSystemQuitHook() {
	installSystemQuitHookOnce.Do(func() {
		C.installOrcaSystemQuitHook()
	})
}

//export OrcaMarkSystemQuit
func OrcaMarkSystemQuit() {
	markSystemQuitRequested()
}

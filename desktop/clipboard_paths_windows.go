//go:build windows

package main

import (
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

const cfHDrop = 15

func readClipboardFilePaths() ([]string, error) {
	user32 := syscall.NewLazyDLL("user32.dll")
	open := user32.NewProc("OpenClipboard")
	close := user32.NewProc("CloseClipboard")
	get := user32.NewProc("GetClipboardData")
	shell32 := syscall.NewLazyDLL("shell32.dll")
	dragQuery := shell32.NewProc("DragQueryFileW")
	if r, _, err := open.Call(0); r == 0 {
		return nil, err
	}
	defer close.Call()
	h, _, err := get.Call(cfHDrop)
	if h == 0 {
		return nil, err
	}
	count, _, _ := dragQuery.Call(h, 0xffffffff, 0, 0)
	paths := make([]string, 0, count)
	for i := uintptr(0); i < count; i++ {
		length, _, _ := dragQuery.Call(h, i, 0, 0)
		buf := make([]uint16, length+1)
		dragQuery.Call(h, i, uintptr(unsafe.Pointer(&buf[0])), length+1)
		paths = append(paths, windows.UTF16ToString(buf))
	}
	return paths, nil
}

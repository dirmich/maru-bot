//go:build windows
package main

import (
	"fmt"
	"syscall"
	"unsafe"
)

func showNativeConfirmDialog(title, message string) bool {
	// Use native Win32 MessageBoxW for maximum reliability
	// 0x00000004 = MB_YESNO, 0x00000020 = MB_ICONQUESTION, 0x00000100 = MB_DEFBUTTON1
	ret := win32MessageBox(message, title, 0x00000004|0x00000020|0x00000100)
	return ret == 6 // 6 = IDYES
}

func showNativeMessageDialog(title, message string) {
	// 0x00000000 = MB_OK, 0x00000040 = MB_ICONINFORMATION
	win32MessageBox(message, title, 0x00000000|0x00000040)
}

func win32MessageBox(message, title string, style uint32) int {
	user32 := syscall.NewLazyDLL("user32.dll")
	proc := user32.NewProc("MessageBoxW")
	lpCaption, _ := syscall.UTF16PtrFromString(title)
	lpText, _ := syscall.UTF16PtrFromString(message)
	ret, _, _ := proc.Call(0, uintptr(unsafe.Pointer(lpText)), uintptr(unsafe.Pointer(lpCaption)), uintptr(style))
	return int(ret)
}

//go:build !windows

package ui

import (
	"os"
	"syscall"
	"unsafe"
)

type winsize struct {
	Row    uint16
	Col    uint16
	Xpixel uint16
	Ypixel uint16
}

func getTerminalWidth() int {
	ws := &winsize{}
	retCode, _, _ := syscall.Syscall(
		syscall.SYS_IOCTL,
		os.Stdout.Fd(),
		uintptr(syscall.TIOCGWINSZ),
		uintptr(unsafe.Pointer(ws)),
	)
	if int(retCode) == 0 && ws.Col > 0 {
		return int(ws.Col)
	}
	return 0
}

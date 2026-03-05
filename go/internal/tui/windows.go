//go:build windows

package tui

import (
	"golang.org/x/sys/windows"
)

const cpUTF8 = 65001

func init() {
	// Force UTF-8 console output on Windows
	_ = windows.SetConsoleOutputCP(cpUTF8)
	_ = windows.SetConsoleCP(cpUTF8)
}

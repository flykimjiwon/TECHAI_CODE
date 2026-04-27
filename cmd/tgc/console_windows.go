//go:build windows

package main

import (
	"os"

	"golang.org/x/sys/windows"
)

// enableWindowsMouse adds ENABLE_MOUSE_INPUT to the console input mode
// so that legacy consoles (conhost.exe) forward mouse events to the app.
// Windows Terminal already supports this via VT input, but this ensures
// compatibility with older Windows environments.
func enableWindowsMouse() {
	h := windows.Handle(os.Stdin.Fd())
	var mode uint32
	if err := windows.GetConsoleMode(h, &mode); err != nil {
		return
	}
	// Add mouse input, remove quick-edit (which intercepts mouse for selection)
	mode |= windows.ENABLE_MOUSE_INPUT
	mode &^= windows.ENABLE_QUICK_EDIT_MODE
	_ = windows.SetConsoleMode(h, mode)
}

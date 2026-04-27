//go:build !windows

package main

func enableWindowsMouse() {
	// No-op on non-Windows platforms.
}

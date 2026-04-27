package main

import (
	"encoding/json"
	"runtime"

	wailsRuntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

// emitEvent sends an event to the frontend.
// macOS: Wails EventsEmit, Windows: SSE broadcast.
func (a *App) emitEvent(name string, data ...interface{}) {
	if runtime.GOOS == "windows" || a.ctx == nil {
		// Browser mode: SSE
		var payload string
		if len(data) > 0 {
			if s, ok := data[0].(string); ok {
				payload = s
			} else {
				b, _ := json.Marshal(data[0])
				payload = string(b)
			}
		}
		sseEmit(name, payload)
	} else {
		// Wails mode
		wailsRuntime.EventsEmit(a.ctx, name, data...)
	}
}

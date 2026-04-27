// Copyright 2025-2026 Kim Jiwon (flykimjiwon). All rights reserved.
// TECHAI CODE — github.com/flykimjiwon/TECHAI_CODE
// Forked from Hanimo Code: github.com/flykimjiwon/hanimo

package main

import (
	"encoding/json"

	wailsRuntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

// emitEvent sends an event to the frontend.
// Electron/browser (ctx==nil): SSE broadcast, Wails (ctx!=nil): native EventsEmit.
func (a *App) emitEvent(name string, data ...interface{}) {
	if a.ctx == nil {
		// Electron / browser mode: SSE
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

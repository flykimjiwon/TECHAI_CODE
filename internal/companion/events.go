package companion

import (
	"encoding/json"
	"fmt"
	"time"
)

// Event types emitted by the TUI and consumed by the browser dashboard.
const (
	EventStateSnapshot = "state_snapshot"
	EventUserMessage   = "user_message"
	EventStreamStart   = "stream_start"
	EventStreamChunk   = "stream_chunk"
	EventStreamDone    = "stream_done"
	EventToolCallStart = "tool_call_start"
	EventToolResult    = "tool_result"
	EventModeChange    = "mode_change"
	EventAutoToggle    = "auto_toggle"
	EventMultiProgress = "multi_progress"
	EventMultiResult   = "multi_result"
	EventSessionCreate = "session_create"
	EventSessionLoad   = "session_load"
)

// Event is a single activity entry stored in the ring buffer and
// streamed to browser clients via SSE.
type Event struct {
	Type      string      `json:"type"`
	Data      interface{} `json:"data,omitempty"`
	ID        int64       `json:"id"`
	Timestamp int64       `json:"ts"`
}

// NewEvent creates a timestamped event. The caller must assign ID
// (the Hub does this automatically via Emit).
func NewEvent(eventType string, data interface{}) Event {
	return Event{
		Type:      eventType,
		Data:      data,
		Timestamp: time.Now().UnixMilli(),
	}
}

// ToSSE serialises the event into the Server-Sent Events wire format:
//
//	event: <type>\nid: <id>\ndata: <json>\n\n
func (e Event) ToSSE() []byte {
	payload, err := json.Marshal(e)
	if err != nil {
		payload = []byte(fmt.Sprintf(`{"type":%q,"error":"marshal failed"}`, e.Type))
	}
	// Standard SSE fields: event, id, data.
	return []byte(fmt.Sprintf("event: %s\nid: %d\ndata: %s\n\n", e.Type, e.ID, payload))
}

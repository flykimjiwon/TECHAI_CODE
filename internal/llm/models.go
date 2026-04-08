package llm

import (
	"strings"
	"sync"
)

type ModelInfo struct {
	ID          string
	DisplayName string
	Description string
	MaxContext  int // maximum context window in tokens
}

var (
	modelsMu sync.RWMutex
	models   = map[string]ModelInfo{
		"openai/gpt-oss-120b": {
			ID:          "openai/gpt-oss-120b",
			DisplayName: "GPT-OSS-120B",
			Description: "범용 대형 모델 — 슈퍼택가이/플랜 메인",
			MaxContext:  131072,
		},
		"qwen/qwen3-coder-30b-a3b-instruct": {
			ID:          "qwen/qwen3-coder-30b-a3b-instruct",
			DisplayName: "Qwen3-Coder-30B",
			Description: "코딩 특화 모델 — 개발 모드 워커",
			MaxContext:  131072,
		},
		"qwen/qwen3-coder-30b": {
			ID:          "qwen/qwen3-coder-30b",
			DisplayName: "Qwen3-Coder-30B",
			Description: "코딩 특화 모델 — 개발 모드 워커",
			MaxContext:  131072,
		},
	}
)

// GetDisplayName returns a human-friendly display name for a model ID.
// Known models get their registered name; unknown models get auto-formatted.
// e.g. "deepseek/deepseek-v3-0324" → "DEEPSEEK-V3-0324"
func GetDisplayName(modelID string) string {
	modelsMu.RLock()
	info, ok := models[modelID]
	modelsMu.RUnlock()
	if ok {
		return info.DisplayName
	}
	return autoDisplayName(modelID)
}

// GetMaxContext returns the max context for a model ID.
// Unknown models default to 131072 (safe assumption for modern LLMs).
func GetMaxContext(modelID string) int {
	modelsMu.RLock()
	info, ok := models[modelID]
	modelsMu.RUnlock()
	if ok {
		return info.MaxContext
	}
	return 131072
}

// autoDisplayName extracts a readable name from a model ID.
// "provider/model-name" → "MODEL-NAME" (uppercase, provider stripped)
func autoDisplayName(modelID string) string {
	name := modelID
	if idx := strings.LastIndex(name, "/"); idx >= 0 {
		name = name[idx+1:]
	}
	return strings.ToUpper(name)
}

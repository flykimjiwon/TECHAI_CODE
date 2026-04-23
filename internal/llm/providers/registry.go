package providers

import (
	"context"
	"fmt"
	"sync"
)

// ChatRequest is the unified request type for all providers.
type ChatRequest struct {
	Model       string
	Messages    []Message
	Tools       []ToolDef
	Temperature float64
	MaxTokens   int
}

// Message represents a chat message.
type Message struct {
	Role       string
	Content    string
	ToolCalls  []ToolCall
	ToolCallID string
	Name       string
}

// ToolCall represents a tool invocation from the model.
type ToolCall struct {
	ID        string
	Name      string
	Arguments string
}

// ToolDef defines a tool the model can call.
type ToolDef struct {
	Name        string
	Description string
	Parameters  map[string]interface{}
}

// ChatChunk is a streaming response fragment.
type ChatChunk struct {
	Content   string
	ToolCalls []ToolCall
	Done      bool
	Error     error
	Usage     *Usage
}

// Usage holds token usage information.
type Usage struct {
	PromptTokens     int
	CompletionTokens int
	TotalTokens      int
}

// ModelInfo describes a model available from a provider.
type ModelInfo struct {
	ID            string
	DisplayName   string
	Provider      string
	ContextWindow int
	SupportsTools bool
}

// Provider is the interface all LLM providers must implement.
type Provider interface {
	Name() string
	Chat(ctx context.Context, req ChatRequest) (<-chan ChatChunk, error)
	ListModels() ([]ModelInfo, error)
	SupportsTools() bool
}

var (
	mu       sync.RWMutex
	registry = map[string]func(baseURL, apiKey string) Provider{}
)

// Register adds a provider factory to the registry.
func Register(name string, factory func(baseURL, apiKey string) Provider) {
	mu.Lock()
	defer mu.Unlock()
	registry[name] = factory
}

// Get returns a provider instance. Falls back to "openai" if name is unknown.
func Get(name, baseURL, apiKey string) (Provider, error) {
	mu.RLock()
	defer mu.RUnlock()
	factory, ok := registry[name]
	if !ok {
		factory, ok = registry["openai"]
		if !ok {
			return nil, fmt.Errorf("프로바이더 %q을(를) 찾을 수 없습니다", name)
		}
	}
	return factory(baseURL, apiKey), nil
}

// Available returns the names of all registered providers.
func Available() []string {
	mu.RLock()
	defer mu.RUnlock()
	names := make([]string, 0, len(registry))
	for name := range registry {
		names = append(names, name)
	}
	return names
}

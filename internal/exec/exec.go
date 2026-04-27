package exec

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	openai "github.com/sashabaranov/go-openai"

	"github.com/kimjiwon/tgc/internal/config"
	"github.com/kimjiwon/tgc/internal/hooks"
	"github.com/kimjiwon/tgc/internal/llm"
	"github.com/kimjiwon/tgc/internal/tools"
)

// Options configures the headless exec mode.
type Options struct {
	Prompt    string
	Stdin     string // piped stdin content
	Model     string
	Mode      llm.Mode
	MaxTurns  int  // max tool-use loop iterations (default 20)
	Ephemeral bool // don't persist session
}

// Run executes a prompt in headless mode (no TUI).
// Output goes to stdout, errors to stderr.
func Run(cfg config.Config, opts Options) error {
	if opts.Prompt == "" {
		return fmt.Errorf("prompt is required")
	}
	if opts.MaxTurns <= 0 {
		opts.MaxTurns = 20
	}

	// Select model
	model := cfg.Models.Super
	if opts.Model != "" {
		model = opts.Model
	}

	// Create LLM client
	client := llm.NewClient(cfg.API.BaseURL, cfg.API.APIKey)

	// Build system prompt
	systemPrompt := llm.SystemPrompt(opts.Mode)

	// Load project context (.techai.md)
	if data, err := os.ReadFile(".techai.md"); err == nil {
		systemPrompt += "\n\n## Project Context\n" + string(data)
	}

	// Build user message — append stdin if piped
	userMsg := opts.Prompt
	if opts.Stdin != "" {
		userMsg += "\n\n<stdin>\n" + opts.Stdin + "\n</stdin>"
	}

	// Initialize hooks
	hookMgr := hooks.NewManager()
	hookMgr.RunSessionStart("exec-headless")

	// Build conversation
	messages := []openai.ChatCompletionMessage{
		{Role: openai.ChatMessageRoleSystem, Content: systemPrompt},
		{Role: openai.ChatMessageRoleUser, Content: userMsg},
	}

	toolDefs := tools.AllTools()

	// Agentic loop
	for turn := 0; turn < opts.MaxTurns; turn++ {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)

		// Stream response
		ch := client.StreamChat(ctx, model, messages, toolDefs)

		var contentBuf strings.Builder
		var toolCalls []llm.ToolCallInfo

		for chunk := range ch {
			if chunk.Err != nil {
				cancel()
				return fmt.Errorf("LLM error: %w", chunk.Err)
			}
			if chunk.Content != "" {
				contentBuf.WriteString(chunk.Content)
				// Stream to stdout in real-time
				fmt.Print(chunk.Content)
			}
			if chunk.Done {
				toolCalls = chunk.ToolCalls
			}
		}
		cancel()

		// Append assistant message
		assistantMsg := openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleAssistant,
			Content: contentBuf.String(),
		}
		if len(toolCalls) > 0 {
			assistantMsg.ToolCalls = toOpenAIToolCalls(toolCalls)
		}
		messages = append(messages, assistantMsg)

		// No tool calls — we're done
		if len(toolCalls) == 0 {
			if contentBuf.Len() > 0 && !strings.HasSuffix(contentBuf.String(), "\n") {
				fmt.Println()
			}
			break
		}

		// Execute tool calls
		for _, tc := range toolCalls {
			// Pre-hook
			hookResult := hookMgr.RunPreToolUse(tc.Name, tc.Arguments)
			if hookResult == hooks.HookFailAbort {
				fmt.Fprintf(os.Stderr, "[hook] pre_tool_use aborted: %s\n", tc.Name)
				toolResult := "Aborted by pre_tool_use hook"
				messages = append(messages, openai.ChatCompletionMessage{
					Role:       openai.ChatMessageRoleTool,
					Content:    toolResult,
					ToolCallID: tc.ID,
				})
				continue
			}

			// Execute
			result := tools.Execute(tc.Name, tc.Arguments)

			// Post-hook
			hookMgr.RunPostToolUse(tc.Name, tc.Arguments, result)

			messages = append(messages, openai.ChatCompletionMessage{
				Role:       openai.ChatMessageRoleTool,
				Content:    result,
				ToolCallID: tc.ID,
			})
		}
	}

	hookMgr.RunStop("exec-headless")
	return nil
}

func toOpenAIToolCalls(calls []llm.ToolCallInfo) []openai.ToolCall {
	result := make([]openai.ToolCall, len(calls))
	for i, tc := range calls {
		result[i] = openai.ToolCall{
			ID:   tc.ID,
			Type: openai.ToolTypeFunction,
			Function: openai.FunctionCall{
				Name:      tc.Name,
				Arguments: tc.Arguments,
			},
		}
	}
	return result
}

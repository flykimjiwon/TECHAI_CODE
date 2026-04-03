package llm

import (
	"context"
	"errors"
	"io"
	"strings"

	openai "github.com/sashabaranov/go-openai"
)

// ToolCallInfo holds a parsed tool call from the API response.
type ToolCallInfo struct {
	ID        string
	Name      string
	Arguments string
}

// StreamChunk represents one piece of a streaming response.
type StreamChunk struct {
	Content   string
	Done      bool
	Err       error
	ToolCalls []ToolCallInfo // non-nil when AI wants to call tools
}

type Client struct {
	api *openai.Client
}

// NormalizeBaseURL ensures the base URL is in the correct format for go-openai.
func NormalizeBaseURL(url string) string {
	url = strings.TrimRight(url, "/")
	url = strings.TrimSuffix(url, "/chat/completions")
	url = strings.TrimRight(url, "/")
	return url
}

func NewClient(baseURL, apiKey string) *Client {
	config := openai.DefaultConfig(apiKey)
	config.BaseURL = NormalizeBaseURL(baseURL)
	return &Client{
		api: openai.NewClientWithConfig(config),
	}
}

// StreamChat streams a chat completion, optionally with tool definitions.
func (c *Client) StreamChat(ctx context.Context, model string, messages []openai.ChatCompletionMessage, toolDefs []openai.Tool) <-chan StreamChunk {
	ch := make(chan StreamChunk)

	go func() {
		defer close(ch)

		req := openai.ChatCompletionRequest{
			Model:    model,
			Messages: messages,
			Stream:   true,
		}
		if len(toolDefs) > 0 {
			req.Tools = toolDefs
		}

		stream, err := c.api.CreateChatCompletionStream(ctx, req)
		if err != nil {
			ch <- StreamChunk{Err: err, Done: true}
			return
		}
		defer stream.Close()

		// Accumulate tool calls from deltas
		tcMap := make(map[int]*ToolCallInfo)

		for {
			resp, err := stream.Recv()
			if errors.Is(err, io.EOF) {
				// Stream finished — check for accumulated tool calls
				if len(tcMap) > 0 {
					calls := make([]ToolCallInfo, 0, len(tcMap))
					for i := 0; i < len(tcMap); i++ {
						if tc, ok := tcMap[i]; ok {
							calls = append(calls, *tc)
						}
					}
					ch <- StreamChunk{Done: true, ToolCalls: calls}
				} else {
					ch <- StreamChunk{Done: true}
				}
				return
			}
			if err != nil {
				ch <- StreamChunk{Err: err, Done: true}
				return
			}

			if len(resp.Choices) == 0 {
				continue
			}

			delta := resp.Choices[0].Delta

			// Stream text content
			if delta.Content != "" {
				ch <- StreamChunk{Content: delta.Content}
			}

			// Accumulate tool call deltas
			for _, tc := range delta.ToolCalls {
				idx := 0
				if tc.Index != nil {
					idx = *tc.Index
				}
				if _, ok := tcMap[idx]; !ok {
					tcMap[idx] = &ToolCallInfo{
						ID:   tc.ID,
						Name: tc.Function.Name,
					}
				} else {
					if tc.ID != "" {
						tcMap[idx].ID = tc.ID
					}
					if tc.Function.Name != "" {
						tcMap[idx].Name = tc.Function.Name
					}
				}
				tcMap[idx].Arguments += tc.Function.Arguments
			}
		}
	}()

	return ch
}

func (c *Client) Chat(ctx context.Context, model string, messages []openai.ChatCompletionMessage) (string, error) {
	resp, err := c.api.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model:    model,
		Messages: messages,
	})
	if err != nil {
		return "", err
	}
	if len(resp.Choices) == 0 {
		return "", errors.New("no response choices")
	}
	return resp.Choices[0].Message.Content, nil
}

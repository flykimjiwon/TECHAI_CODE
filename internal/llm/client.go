package llm

import (
	"context"
	"errors"
	"io"
	"strings"

	openai "github.com/sashabaranov/go-openai"
)

type StreamChunk struct {
	Content string
	Done    bool
	Err     error
}

type Client struct {
	api *openai.Client
}

// NormalizeBaseURL ensures the base URL is in the correct format for go-openai.
// The library appends /chat/completions automatically, so:
//   - "https://api.novita.ai/openai/chat/completions" → "https://api.novita.ai/openai"
//   - "https://api.openai.com/v1/" → "https://api.openai.com/v1"
//   - "https://api.example.com" → "https://api.example.com"
func NormalizeBaseURL(url string) string {
	url = strings.TrimRight(url, "/")
	// Strip /chat/completions suffix if user pasted the full endpoint
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

func (c *Client) StreamChat(ctx context.Context, model string, messages []openai.ChatCompletionMessage) <-chan StreamChunk {
	ch := make(chan StreamChunk)

	go func() {
		defer close(ch)

		stream, err := c.api.CreateChatCompletionStream(ctx, openai.ChatCompletionRequest{
			Model:    model,
			Messages: messages,
			Stream:   true,
		})
		if err != nil {
			ch <- StreamChunk{Err: err, Done: true}
			return
		}
		defer stream.Close()

		for {
			resp, err := stream.Recv()
			if errors.Is(err, io.EOF) {
				ch <- StreamChunk{Done: true}
				return
			}
			if err != nil {
				ch <- StreamChunk{Err: err, Done: true}
				return
			}
			if len(resp.Choices) > 0 {
				delta := resp.Choices[0].Delta.Content
				if delta != "" {
					ch <- StreamChunk{Content: delta}
				}
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

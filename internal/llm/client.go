package llm

import (
	"context"
	"errors"
	"io"
	"net/http"
	"os"
	"strings"

	openai "github.com/sashabaranov/go-openai"

	"github.com/kimjiwon/tgc/internal/config"
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
	cfg := openai.DefaultConfig(apiKey)
	cfg.BaseURL = NormalizeBaseURL(baseURL)

	// Log network environment for debug builds
	if config.IsDebug() {
		config.DebugLog("[NET] HTTP_PROXY=%s", os.Getenv("HTTP_PROXY"))
		config.DebugLog("[NET] HTTPS_PROXY=%s", os.Getenv("HTTPS_PROXY"))
		config.DebugLog("[NET] NO_PROXY=%s", os.Getenv("NO_PROXY"))
		config.DebugLog("[NET] API BaseURL=%s", cfg.BaseURL)

		// Wrap transport to log TLS and response headers
		origTransport := http.DefaultTransport.(*http.Transport).Clone()
		cfg.HTTPClient = &http.Client{
			Transport: &debugTransport{inner: origTransport},
		}
	}

	return &Client{
		api: openai.NewClientWithConfig(cfg),
	}
}

// debugTransport wraps an http.RoundTripper to log request/response details.
type debugTransport struct {
	inner http.RoundTripper
}

func (d *debugTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	config.DebugLog("[NET-REQ] %s %s | Content-Type=%s | Accept=%s",
		req.Method, req.URL.String(),
		req.Header.Get("Content-Type"),
		req.Header.Get("Accept"))

	resp, err := d.inner.RoundTrip(req)
	if err != nil {
		config.DebugLog("[NET-ERR] %v", err)
		return resp, err
	}

	config.DebugLog("[NET-RES] Status=%d | Content-Type=%s | Transfer-Encoding=%v",
		resp.StatusCode, resp.Header.Get("Content-Type"), resp.TransferEncoding)

	if resp.TLS != nil && len(resp.TLS.PeerCertificates) > 0 {
		cert := resp.TLS.PeerCertificates[0]
		config.DebugLog("[NET-TLS] server cert issuer=%s | subject=%s",
			cert.Issuer.CommonName, cert.Subject.CommonName)
	}

	return resp, nil
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

		config.DebugLog("[API-REQ] POST /chat/completions | model=%s | msgs=%d | tools=%d", model, len(messages), len(toolDefs))

		stream, err := c.api.CreateChatCompletionStream(ctx, req)
		if err != nil {
			config.DebugLog("[API-ERR] stream create failed: %v", err)
			ch <- StreamChunk{Err: err, Done: true}
			return
		}
		defer stream.Close()

		config.DebugLog("[API-RES] stream opened successfully")

		// Accumulate tool calls from deltas
		tcMap := make(map[int]*ToolCallInfo)
		chunkNum := 0
		totalContentLen := 0

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
					config.DebugLog("[STREAM-DONE] chunks=%d | totalContent=%dbytes | toolCalls=%d", chunkNum, totalContentLen, len(calls))
					for i, tc := range calls {
						config.DebugLog("[STREAM-DONE] toolCall[%d] name=%s | argsLen=%d", i, tc.Name, len(tc.Arguments))
					}
					ch <- StreamChunk{Done: true, ToolCalls: calls}
				} else {
					config.DebugLog("[STREAM-DONE] chunks=%d | totalContent=%dbytes | toolCalls=0", chunkNum, totalContentLen)
					ch <- StreamChunk{Done: true}
				}
				return
			}
			if err != nil {
				config.DebugLog("[STREAM-ERR] after %d chunks: %v", chunkNum, err)
				ch <- StreamChunk{Err: err, Done: true}
				return
			}

			chunkNum++

			if len(resp.Choices) == 0 {
				config.DebugLog("[CHUNK#%d] empty choices", chunkNum)
				continue
			}

			delta := resp.Choices[0].Delta

			// Stream text content
			if delta.Content != "" {
				totalContentLen += len(delta.Content)
				hasTC := len(delta.ToolCalls) > 0
				config.DebugLog("[CHUNK#%d] content len=%d | toolCall=%v", chunkNum, len(delta.Content), hasTC)
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
					config.DebugLog("[CHUNK#%d] toolCall[%d] START name=%s | id=%s", chunkNum, idx, tc.Function.Name, tc.ID)
				} else {
					if tc.ID != "" {
						tcMap[idx].ID = tc.ID
					}
					if tc.Function.Name != "" {
						tcMap[idx].Name = tc.Function.Name
					}
				}
				tcMap[idx].Arguments += tc.Function.Arguments
				if len(tc.Function.Arguments) > 0 {
					config.DebugLog("[CHUNK#%d] toolCall[%d] argsDelta len=%d", chunkNum, idx, len(tc.Function.Arguments))
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

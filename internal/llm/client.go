package llm

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

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
	Content    string
	IsThinking bool   // true when content came from reasoning_content (thinking phase)
	Done       bool
	Err        error
	ToolCalls  []ToolCallInfo // non-nil when AI wants to call tools
	// Usage is populated on the final Done chunk when the provider supports
	// stream_options.include_usage (OpenAI, most compatible gateways). It is
	// nil when the provider does not return usage statistics, in which case
	// callers should fall back to a chunk-based or char-based estimate.
	Usage *openai.Usage
}

// tokenCountFromUsage extracts a usable token total from an optional
// *openai.Usage. It is nil-safe and falls back to prompt+completion when
// TotalTokens is 0 (some compatible providers leave TotalTokens unset).
func tokenCountFromUsage(u *openai.Usage) int {
	if u == nil {
		return 0
	}
	if u.TotalTokens > 0 {
		return u.TotalTokens
	}
	return u.PromptTokens + u.CompletionTokens
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
		config.DebugLog("[NET] API BaseURL=%s", cfg.BaseURL)
		config.DebugLog("[NET] API Key length=%d prefix=%s", len(apiKey), safePrefix(apiKey, 8))

		// DNS pre-check
		host := extractHost(cfg.BaseURL)
		if host != "" {
			config.DebugLog("[NET-DNS] resolving %s ...", host)
			start := time.Now()
			addrs, err := net.LookupHost(host)
			elapsed := time.Since(start)
			if err != nil {
				config.DebugLog("[NET-DNS] FAILED %s: %v (took %v)", host, err, elapsed)
			} else {
				config.DebugLog("[NET-DNS] OK %s → %v (took %v)", host, addrs, elapsed)
			}
		}

		// NOTE: debugTransport HTTP wrapping removed — it blocks SSE streaming.
		// See docs/DEBUG_TRANSPORT_FREEZE.md for details.
		// Application-level DebugLog calls in StreamChat are sufficient.
	}

	return &Client{
		api: openai.NewClientWithConfig(cfg),
	}
}

func extractHost(baseURL string) string {
	// Remove scheme
	u := baseURL
	if idx := strings.Index(u, "://"); idx >= 0 {
		u = u[idx+3:]
	}
	// Remove path
	if idx := strings.Index(u, "/"); idx >= 0 {
		u = u[:idx]
	}
	// Remove port
	if host, _, err := net.SplitHostPort(u); err == nil {
		return host
	}
	return u
}

func safePrefix(s string, n int) string {
	if len(s) <= n {
		return s + "..."
	}
	return s[:n] + "..."
}

// debugTransport wraps an http.RoundTripper to log request/response details.
type debugTransport struct {
	inner    http.RoundTripper
	reqCount int
}

func (d *debugTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	d.reqCount++
	reqID := d.reqCount

	// Log all request headers
	config.DebugLog("[NET-REQ#%d] %s %s", reqID, req.Method, req.URL.String())
	for k, v := range req.Header {
		config.DebugLog("[NET-REQ#%d] Header: %s=%s", reqID, k, strings.Join(v, ", "))
	}
	if req.ContentLength > 0 {
		config.DebugLog("[NET-REQ#%d] ContentLength=%d", reqID, req.ContentLength)
	}

	start := time.Now()
	resp, err := d.inner.RoundTrip(req)
	elapsed := time.Since(start)

	if err != nil {
		config.DebugLog("[NET-ERR#%d] after %v: %v", reqID, elapsed, err)
		// Classify error type
		if netErr, ok := err.(net.Error); ok {
			config.DebugLog("[NET-ERR#%d] timeout=%v | temporary=%v", reqID, netErr.Timeout(), false)
		}
		if os.IsTimeout(err) {
			config.DebugLog("[NET-ERR#%d] OS timeout detected", reqID)
		}
		return resp, err
	}

	config.DebugLog("[NET-RES#%d] Status=%d | elapsed=%v", reqID, resp.StatusCode, elapsed)
	config.DebugLog("[NET-RES#%d] Content-Type=%s", reqID, resp.Header.Get("Content-Type"))
	config.DebugLog("[NET-RES#%d] Transfer-Encoding=%v", reqID, resp.TransferEncoding)
	config.DebugLog("[NET-RES#%d] Content-Length=%s", reqID, resp.Header.Get("Content-Length"))
	config.DebugLog("[NET-RES#%d] Connection=%s", reqID, resp.Header.Get("Connection"))

	// Log all response headers for full visibility
	for k, v := range resp.Header {
		config.DebugLog("[NET-RES#%d] Header: %s=%s", reqID, k, strings.Join(v, ", "))
	}

	// TLS details
	if resp.TLS != nil {
		config.DebugLog("[NET-TLS#%d] Version=0x%04x CipherSuite=0x%04x", reqID, resp.TLS.Version, resp.TLS.CipherSuite)
		config.DebugLog("[NET-TLS#%d] ServerName=%s", reqID, resp.TLS.ServerName)
		for i, cert := range resp.TLS.PeerCertificates {
			config.DebugLog("[NET-TLS#%d] Cert[%d] Subject=%s | Issuer=%s | NotAfter=%s",
				reqID, i, cert.Subject.CommonName, cert.Issuer.CommonName, cert.NotAfter.Format("2006-01-02"))
		}
	} else {
		config.DebugLog("[NET-TLS#%d] NO TLS (plain HTTP or proxy terminated)", reqID)
	}

	// Detect proxy
	if via := resp.Header.Get("Via"); via != "" {
		config.DebugLog("[NET-PROXY#%d] Via: %s", reqID, via)
	}
	if xfwd := resp.Header.Get("X-Forwarded-For"); xfwd != "" {
		config.DebugLog("[NET-PROXY#%d] X-Forwarded-For: %s", reqID, xfwd)
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
			// Ask the provider to include a final usage chunk so we can
			// update the HUD token counter with real prompt+completion
			// totals instead of the chunk-count fallback.
			StreamOptions: &openai.StreamOptions{
				IncludeUsage: true,
			},
		}
		if len(toolDefs) > 0 {
			req.Tools = toolDefs
		}

		// Log tool definitions being sent
		config.DebugLog("[API-REQ] POST /chat/completions | model=%s | msgs=%d | tools=%d", model, len(messages), len(toolDefs))
		for i, td := range toolDefs {
			if td.Function != nil {
				config.DebugLog("[API-REQ] tool[%d] name=%s", i, td.Function.Name)
			}
		}
		// Log last message role/content preview
		if len(messages) > 0 {
			last := messages[len(messages)-1]
			preview := last.Content
			if len(preview) > 200 {
				preview = preview[:200] + "..."
			}
			config.DebugLog("[API-REQ] lastMsg role=%s | len=%d | preview=%q", last.Role, len(last.Content), preview)
		}

		// Serialize request to see what's actually sent
		if config.IsDebug() {
			reqJSON, _ := json.Marshal(req)
			config.DebugLog("[API-REQ-RAW] len=%d", len(reqJSON))
			if len(reqJSON) < 5000 {
				config.DebugLog("[API-REQ-RAW] %s", string(reqJSON))
			} else {
				config.DebugLog("[API-REQ-RAW] (truncated) %s", string(reqJSON[:5000]))
			}
		}

		streamStart := time.Now()
		stream, err := c.api.CreateChatCompletionStream(ctx, req)
		streamOpenElapsed := time.Since(streamStart)
		if err != nil {
			config.DebugLog("[API-ERR] stream create failed after %v: %v", streamOpenElapsed, err)
			ch <- StreamChunk{Err: err, Done: true}
			return
		}
		defer stream.Close()

		config.DebugLog("[API-RES] stream opened in %v", streamOpenElapsed)

		// Accumulate tool calls from deltas
		tcMap := make(map[int]*ToolCallInfo)
		var fullContentBuf strings.Builder // tracks full content for text tool_call parsing
		chunkNum := 0
		totalContentLen := 0
		lastChunkTime := time.Now()
		firstChunkTime := time.Time{}
		// lastUsage captures the final usage chunk the provider emits when
		// stream_options.include_usage is set. It is propagated on the Done
		// chunk so the HUD can show real token counts.
		var lastUsage *openai.Usage

		for {
			recvStart := time.Now()
			resp, err := stream.Recv()
			recvElapsed := time.Since(recvStart)

			if errors.Is(err, io.EOF) {
				totalElapsed := time.Since(streamStart)
				// Stream finished — check for accumulated tool calls
				if len(tcMap) > 0 {
					calls := make([]ToolCallInfo, 0, len(tcMap))
					for i := 0; i < len(tcMap); i++ {
						if tc, ok := tcMap[i]; ok {
							calls = append(calls, *tc)
						}
					}
					config.DebugLog("[STREAM-DONE] chunks=%d | totalContent=%dbytes | toolCalls=%d | totalTime=%v | usageTokens=%d", chunkNum, totalContentLen, len(calls), totalElapsed, tokenCountFromUsage(lastUsage))
					for i, tc := range calls {
						config.DebugLog("[STREAM-DONE] toolCall[%d] name=%s | argsLen=%d | args=%s", i, tc.Name, len(tc.Arguments), truncateForLog(tc.Arguments, 500))
					}
					ch <- StreamChunk{Done: true, ToolCalls: calls, Usage: lastUsage}
				} else {
					// Fallback: parse <tool_call> tags from streamed content.
					// Some API proxies pass tools to the model but don't convert
					// the model's native tool_call format to OpenAI tool_calls.
					// Qwen3 outputs: <tool_call>{"name":"...","arguments":{...}}</tool_call>
					if parsed := parseToolCallsFromContent(fullContentBuf.String()); len(parsed) > 0 {
						config.DebugLog("[STREAM-DONE] chunks=%d | totalContent=%dbytes | textToolCalls=%d (parsed from content) | totalTime=%v | usageTokens=%d", chunkNum, totalContentLen, len(parsed), totalElapsed, tokenCountFromUsage(lastUsage))
						for i, tc := range parsed {
							config.DebugLog("[STREAM-DONE] textToolCall[%d] name=%s | argsLen=%d", i, tc.Name, len(tc.Arguments))
						}
						ch <- StreamChunk{Done: true, ToolCalls: parsed, Usage: lastUsage}
					} else {
						config.DebugLog("[STREAM-DONE] chunks=%d | totalContent=%dbytes | toolCalls=0 | totalTime=%v | usageTokens=%d", chunkNum, totalContentLen, totalElapsed, tokenCountFromUsage(lastUsage))
						ch <- StreamChunk{Done: true, Usage: lastUsage}
					}
				}
				return
			}
			if err != nil {
				totalElapsed := time.Since(streamStart)
				config.DebugLog("[STREAM-ERR] after %d chunks, %v total: %v", chunkNum, totalElapsed, err)
				config.DebugLog("[STREAM-ERR] lastChunkAge=%v | recvWait=%v", time.Since(lastChunkTime), recvElapsed)
				ch <- StreamChunk{Err: err, Done: true}
				return
			}

			chunkNum++
			now := time.Now()
			gap := now.Sub(lastChunkTime)
			lastChunkTime = now
			if firstChunkTime.IsZero() {
				firstChunkTime = now
				config.DebugLog("[STREAM] first chunk after %v (TTFC)", now.Sub(streamStart))
			}

			// Capture usage from whichever chunk the provider attaches it
			// to. With StreamOptions.IncludeUsage=true, OpenAI emits a
			// final chunk with empty Choices and non-nil Usage; some
			// gateways attach it to the last content chunk instead.
			if resp.Usage != nil {
				lastUsage = resp.Usage
				config.DebugLog("[CHUNK#%d] usage prompt=%d completion=%d total=%d",
					chunkNum, resp.Usage.PromptTokens, resp.Usage.CompletionTokens, resp.Usage.TotalTokens)
			}

			if len(resp.Choices) == 0 {
				config.DebugLog("[CHUNK#%d] empty choices | gap=%v", chunkNum, gap)
				continue
			}

			delta := resp.Choices[0].Delta
			finishReason := ""
			if resp.Choices[0].FinishReason != "" {
				finishReason = string(resp.Choices[0].FinishReason)
				config.DebugLog("[CHUNK#%d] finishReason=%s", chunkNum, finishReason)
			}

			// Stream text content — handle both content and reasoning_content.
			// Some models (e.g. GPT-OSS-120B via Novita) send thinking in
			// reasoning_content first, then the final answer in content.
			if delta.Content != "" {
				totalContentLen += len(delta.Content)
				fullContentBuf.WriteString(delta.Content)
				hasTC := len(delta.ToolCalls) > 0
				if chunkNum <= 5 || chunkNum%50 == 0 {
					config.DebugLog("[CHUNK#%d] content len=%d | toolCall=%v | gap=%v | preview=%q", chunkNum, len(delta.Content), hasTC, gap, truncateForLog(delta.Content, 100))
				}
				ch <- StreamChunk{Content: delta.Content, IsThinking: false}
			} else if delta.ReasoningContent != "" {
				totalContentLen += len(delta.ReasoningContent)
				if chunkNum <= 5 || chunkNum%50 == 0 {
					config.DebugLog("[CHUNK#%d] reasoning len=%d | gap=%v | preview=%q", chunkNum, len(delta.ReasoningContent), gap, truncateForLog(delta.ReasoningContent, 100))
				}
				ch <- StreamChunk{Content: delta.ReasoningContent, IsThinking: true}
			} else if len(delta.ToolCalls) == 0 && delta.Role == "" {
				config.DebugLog("[CHUNK#%d] EMPTY (no content, no toolCalls) | gap=%v", chunkNum, gap)
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
					config.DebugLog("[CHUNK#%d] toolCall[%d] START name=%q | id=%q | gap=%v", chunkNum, idx, tc.Function.Name, tc.ID, gap)
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
					config.DebugLog("[CHUNK#%d] toolCall[%d] argsDelta len=%d | accumulated=%d", chunkNum, idx, len(tc.Function.Arguments), len(tcMap[idx].Arguments))
				}
			}
		}
	}()

	return ch
}

func truncateForLog(s string, max int) string {
	s = strings.ReplaceAll(s, "\n", "\\n")
	if len(s) > max {
		return s[:max] + "..."
	}
	return s
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
	content := resp.Choices[0].Message.Content
	// Fallback to reasoning_content if content is empty (e.g. GPT-OSS-120B)
	if content == "" && resp.Choices[0].Message.ReasoningContent != "" {
		content = resp.Choices[0].Message.ReasoningContent
	}
	return content, nil
}

// parseToolCallsFromContent extracts tool calls from text content when
// the API proxy doesn't convert model-native tool call format to OpenAI.
// Supports Qwen3 format: <tool_call>{"name":"...","arguments":{...}}</tool_call>
// and also: <function=name>{"key":"val"}</function>
func parseToolCallsFromContent(content string) []ToolCallInfo {
	var calls []ToolCallInfo

	// Pattern 1: <tool_call>{"name":"...","arguments":{...}}</tool_call>
	for {
		start := strings.Index(content, "<tool_call>")
		if start == -1 {
			break
		}
		end := strings.Index(content[start:], "</tool_call>")
		if end == -1 {
			break
		}
		jsonStr := strings.TrimSpace(content[start+len("<tool_call>") : start+end])
		content = content[start+end+len("</tool_call>"):]

		var parsed struct {
			Name      string          `json:"name"`
			Arguments json.RawMessage `json:"arguments"`
		}
		if err := json.Unmarshal([]byte(jsonStr), &parsed); err != nil {
			config.DebugLog("[TOOL-PARSE] failed to parse tool_call JSON: %v", err)
			continue
		}
		calls = append(calls, ToolCallInfo{
			ID:        fmt.Sprintf("text-tc-%d", len(calls)),
			Name:      parsed.Name,
			Arguments: string(parsed.Arguments),
		})
	}

	// Pattern 2: <function=file_read>{"path":"..."}</function>
	remaining := content
	for {
		funcStart := strings.Index(remaining, "<function=")
		if funcStart == -1 {
			break
		}
		nameEnd := strings.Index(remaining[funcStart:], ">")
		if nameEnd == -1 {
			break
		}
		name := remaining[funcStart+len("<function=") : funcStart+nameEnd]

		funcEnd := strings.Index(remaining[funcStart:], "</function>")
		if funcEnd == -1 {
			// Also try </tool_call> as some models mix formats
			funcEnd = strings.Index(remaining[funcStart:], "</tool_call>")
			if funcEnd == -1 {
				break
			}
		}
		argsStr := strings.TrimSpace(remaining[funcStart+nameEnd+1 : funcStart+funcEnd])
		remaining = remaining[funcStart+funcEnd+len("</function>"):]

		// Validate it's JSON
		if len(argsStr) > 0 && argsStr[0] == '{' {
			calls = append(calls, ToolCallInfo{
				ID:        fmt.Sprintf("text-tc-%d", len(calls)),
				Name:      name,
				Arguments: argsStr,
			})
		}
	}

	return calls
}

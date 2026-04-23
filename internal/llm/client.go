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
		var sentContentLen int             // bytes of fullContentBuf already sent as StreamChunks
		var toolCallOutputStarted bool     // true once <tool_call>/<function= detected in content
		var insideThinkTag bool            // true while inside <think>...</think> in content
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
					// Also check for additional text tool calls the proxy
					// may not have converted (partial proxy conversion).
					if textCalls := parseToolCallsFromContent(fullContentBuf.String()); len(textCalls) > 0 {
						existing := make(map[string]bool)
						for _, c := range calls {
							existing[c.Name+"|"+c.Arguments] = true
						}
						for _, tc := range textCalls {
							if !existing[tc.Name+"|"+tc.Arguments] {
								tc.ID = fmt.Sprintf("text-tc-%d", len(calls))
								calls = append(calls, tc)
								config.DebugLog("[STREAM-DONE] merged text tool call: name=%s", tc.Name)
							}
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
						// No tool calls — flush any held-back partial tag suffix
						// as regular content (it wasn't a real tag after all).
						full := fullContentBuf.String()
						if sentContentLen < len(full) {
							remainder := full[sentContentLen:]
							for _, seg := range splitThinkSegments(remainder, &insideThinkTag) {
								ch <- seg
							}
						}
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
			//
			// Tool-call tag suppression: when the model outputs tool calls as
			// text (e.g. <tool_call>JSON</tool_call>), we detect the tag start
			// and stop forwarding content to the UI. The raw text is still
			// accumulated in fullContentBuf for parsing at EOF.
			if delta.Content != "" {
				totalContentLen += len(delta.Content)
				fullContentBuf.WriteString(delta.Content)

				if !toolCallOutputStarted {
					full := fullContentBuf.String()
					tagIdx := findToolCallTagStart(full)

					if tagIdx >= 0 {
						toolCallOutputStarted = true
						// Send only the content before the tag that hasn't been sent yet
						if tagIdx > sentContentLen {
							unsent := full[sentContentLen:tagIdx]
							for _, seg := range splitThinkSegments(unsent, &insideThinkTag) {
								ch <- seg
							}
						}
						config.DebugLog("[CHUNK#%d] tool_call tag detected at pos %d, suppressing output", chunkNum, tagIdx)
					} else {
						// Check for partial tag at the end of accumulated content.
						// E.g. "<function " could be the start of "<function=name>"
						// split across chunks. Hold back the partial suffix.
						safeEnd := len(full)
						if hold := partialToolTagSuffix(full); hold > 0 {
							safeEnd = len(full) - hold
						}
						if safeEnd > sentContentLen {
							unsent := full[sentContentLen:safeEnd]
							for _, seg := range splitThinkSegments(unsent, &insideThinkTag) {
								ch <- seg
							}
						}
						sentContentLen = safeEnd
					}
				}
				// When toolCallOutputStarted: content goes to fullContentBuf only, not to UI

				hasTC := len(delta.ToolCalls) > 0
				if chunkNum <= 5 || chunkNum%50 == 0 {
					config.DebugLog("[CHUNK#%d] content len=%d | toolCall=%v | gap=%v | suppressed=%v | thinking=%v | preview=%q", chunkNum, len(delta.Content), hasTC, gap, toolCallOutputStarted, insideThinkTag, truncateForLog(delta.Content, 100))
				}
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
//
// Supported patterns (Qwen3-Coder / ChatGLM / other OSS models):
//   - <tool_call>{"name":"...","arguments":{...}}</tool_call>
//   - <|tool_call|>{"name":"...","arguments":{...}}<|/tool_call|>
//   - <function=name>{"key":"val"}</function>
//   - <function=name>{"key":"val"}</tool_call>
//
// Thinking blocks (<think>...</think>) are stripped before parsing to
// prevent false-positive extraction from model reasoning.
func parseToolCallsFromContent(content string) []ToolCallInfo {
	// ── Step 0: strip <think>...</think> blocks ──
	content = stripThinkTags(content)
	if content == "" {
		return nil
	}

	var calls []ToolCallInfo

	// ── Pattern 1a: <tool_call>...</tool_call> ──
	content = parseToolCallTag(content, "<tool_call>", "</tool_call>", &calls)

	// ── Pattern 1b: <|tool_call|>...<|/tool_call|> (pipe-delimited variant) ──
	content = parseToolCallTag(content, "<|tool_call|>", "<|/tool_call|>", &calls)

	// ── Pattern 2: <function=name>{...}</function> or </tool_call> ──
	remaining := content
	for {
		funcStart := strings.Index(remaining, "<function=")
		if funcStart == -1 {
			break
		}
		// Find closing '>' of opening tag
		nameEnd := strings.Index(remaining[funcStart:], ">")
		if nameEnd == -1 {
			break
		}
		name := strings.TrimSpace(remaining[funcStart+len("<function=") : funcStart+nameEnd])
		if name == "" {
			// Empty function name — skip malformed tag
			remaining = remaining[funcStart+nameEnd+1:]
			continue
		}

		// Find closing tag — could be </function> or </tool_call>
		closeTag := "</function>"
		funcEnd := strings.Index(remaining[funcStart:], closeTag)
		if funcEnd == -1 {
			closeTag = "</tool_call>"
			funcEnd = strings.Index(remaining[funcStart:], closeTag)
			if funcEnd == -1 {
				break
			}
		}
		argsStr := strings.TrimSpace(remaining[funcStart+nameEnd+1 : funcStart+funcEnd])
		remaining = remaining[funcStart+funcEnd+len(closeTag):]

		// Parse arguments — supports both JSON and <parameter=key> format.
		if len(argsStr) == 0 {
			argsStr = "{}"
		}
		if argsStr[0] != '{' {
			// Not JSON — try <parameter=key> value format.
			// Example: <parameter=path> . <parameter=recursive> false
			if params := parseParameterTags(argsStr); len(params) > 0 {
				paramsJSON, err := json.Marshal(params)
				if err == nil {
					argsStr = string(paramsJSON)
					config.DebugLog("[TOOL-PARSE] function=%s: converted %d parameter tags to JSON", name, len(params))
				} else {
					config.DebugLog("[TOOL-PARSE] function=%s: parameter tags marshal failed: %v", name, err)
					continue
				}
			} else {
				config.DebugLog("[TOOL-PARSE] function=%s: args not JSON and no parameter tags: %q", name, truncateForLog(argsStr, 100))
				continue
			}
		}
		if !json.Valid([]byte(argsStr)) {
			config.DebugLog("[TOOL-PARSE] function=%s: invalid JSON args: %q", name, truncateForLog(argsStr, 200))
			continue
		}

		calls = append(calls, ToolCallInfo{
			ID:        fmt.Sprintf("text-tc-%d", len(calls)),
			Name:      name,
			Arguments: argsStr,
		})
	}

	return calls
}

// parseToolCallTag extracts tool calls matching a specific open/close tag pair.
// The remaining content (after all matched tags are consumed) is returned.
func parseToolCallTag(content, openTag, closeTag string, calls *[]ToolCallInfo) string {
	for {
		start := strings.Index(content, openTag)
		if start == -1 {
			break
		}
		end := strings.Index(content[start:], closeTag)
		if end == -1 {
			// Unclosed tag — model was likely cut off (finishReason=length).
			// Try to parse what we have; if JSON is also truncated, log warning.
			jsonStr := strings.TrimSpace(content[start+len(openTag):])
			if len(jsonStr) > 0 && jsonStr[0] == '{' {
				if tc, ok := parseToolCallJSON(jsonStr, len(*calls)); ok {
					config.DebugLog("[TOOL-PARSE] recovered unclosed %s tag: name=%s", openTag, tc.Name)
					*calls = append(*calls, tc)
				} else {
					config.DebugLog("[TOOL-PARSE] unclosed %s with truncated JSON (context overflow?): %q", openTag, truncateForLog(jsonStr, 200))
				}
			} else {
				config.DebugLog("[TOOL-PARSE] unclosed %s with no valid JSON content", openTag)
			}
			content = ""
			break
		}
		jsonStr := strings.TrimSpace(content[start+len(openTag) : start+end])
		content = content[start+end+len(closeTag):]

		if tc, ok := parseToolCallJSON(jsonStr, len(*calls)); ok {
			*calls = append(*calls, tc)
		}
	}
	return content
}

// parseToolCallJSON parses a JSON string from a tool_call tag into a ToolCallInfo.
// Returns (info, true) on success, (zero, false) on failure.
func parseToolCallJSON(jsonStr string, idx int) (ToolCallInfo, bool) {
	if len(jsonStr) == 0 {
		return ToolCallInfo{}, false
	}
	// Some models prepend text before JSON: "Sure!\n{"name":"..."}
	// Find the first '{' and try from there.
	if jsonStr[0] != '{' {
		braceIdx := strings.Index(jsonStr, "{")
		if braceIdx == -1 {
			config.DebugLog("[TOOL-PARSE] no JSON object found: %q", truncateForLog(jsonStr, 100))
			return ToolCallInfo{}, false
		}
		config.DebugLog("[TOOL-PARSE] skipping %d bytes of preamble before JSON", braceIdx)
		jsonStr = jsonStr[braceIdx:]
	}

	// Try standard format: {"name":"...", "arguments":{...}}
	var parsed struct {
		Name      string          `json:"name"`
		Arguments json.RawMessage `json:"arguments"`
	}
	if err := json.Unmarshal([]byte(jsonStr), &parsed); err != nil {
		config.DebugLog("[TOOL-PARSE] JSON unmarshal failed: %v | raw=%q", err, truncateForLog(jsonStr, 200))
		return ToolCallInfo{}, false
	}
	if parsed.Name == "" {
		config.DebugLog("[TOOL-PARSE] empty tool name in JSON: %q", truncateForLog(jsonStr, 100))
		return ToolCallInfo{}, false
	}

	args := string(parsed.Arguments)
	if args == "" || args == "null" {
		args = "{}"
	}

	// Handle arguments as escaped JSON string.
	// Some models output: {"name":"x","arguments":"{\"path\":\"y\"}"}
	// instead of:         {"name":"x","arguments":{"path":"y"}}
	if len(args) >= 2 && args[0] == '"' {
		var unescaped string
		if err := json.Unmarshal([]byte(args), &unescaped); err == nil && len(unescaped) > 0 && unescaped[0] == '{' {
			args = unescaped
		}
	}

	return ToolCallInfo{
		ID:        fmt.Sprintf("text-tc-%d", idx),
		Name:      parsed.Name,
		Arguments: args,
	}, true
}

// parseParameterTags parses <parameter=key> value sequences into a map.
// This format is used by some Qwen3 variants instead of JSON arguments:
//
//	<parameter=path> . <parameter=recursive> false
//	→ {"path": ".", "recursive": "false"}
//
// Also handles the explicit closing form: <parameter=key>value</parameter>
func parseParameterTags(s string) map[string]string {
	params := make(map[string]string)
	parts := strings.Split(s, "<parameter=")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		// part looks like: "path> ." or "recursive> false" or "path>./src</parameter>"
		closeIdx := strings.Index(part, ">")
		if closeIdx == -1 {
			continue
		}
		key := strings.TrimSpace(part[:closeIdx])
		value := strings.TrimSpace(part[closeIdx+1:])
		// Remove trailing </parameter> if present
		if idx := strings.Index(value, "</parameter>"); idx >= 0 {
			value = strings.TrimSpace(value[:idx])
		}
		if key != "" && value != "" {
			params[key] = value
		}
	}
	return params
}

// stripThinkTags removes <think>...</think> blocks from content.
// Handles unclosed tags (strips from <think> to end) and nested tags.
func stripThinkTags(s string) string {
	for {
		start := strings.Index(s, "<think>")
		if start == -1 {
			// Also check <|think|> variant
			start = strings.Index(s, "<|think|>")
			if start == -1 {
				break
			}
			closeTag := "<|/think|>"
			end := strings.Index(s[start:], closeTag)
			if end == -1 {
				s = s[:start]
				break
			}
			s = s[:start] + s[start+end+len(closeTag):]
			continue
		}
		end := strings.Index(s[start:], "</think>")
		if end == -1 {
			// Unclosed — strip from <think> to end
			s = s[:start]
			break
		}
		s = s[:start] + s[start+end+len("</think>"):]
	}
	return s
}

// splitThinkSegments splits content at <think>/<​/think> boundaries and
// returns StreamChunks with correct IsThinking flags. The insideThink
// pointer tracks state across calls (chunk boundaries). The raw tags
// themselves are stripped from output so the user never sees them.
func splitThinkSegments(content string, insideThink *bool) []StreamChunk {
	var segs []StreamChunk
	for len(content) > 0 {
		if *insideThink {
			if idx := strings.Index(content, "</think>"); idx >= 0 {
				if idx > 0 {
					segs = append(segs, StreamChunk{Content: content[:idx], IsThinking: true})
				}
				content = content[idx+len("</think>"):]
				*insideThink = false
			} else {
				segs = append(segs, StreamChunk{Content: content, IsThinking: true})
				content = ""
			}
		} else {
			if idx := strings.Index(content, "<think>"); idx >= 0 {
				if idx > 0 {
					segs = append(segs, StreamChunk{Content: content[:idx], IsThinking: false})
				}
				content = content[idx+len("<think>"):]
				*insideThink = true
			} else {
				segs = append(segs, StreamChunk{Content: content, IsThinking: false})
				content = ""
			}
		}
	}
	return segs
}

// partialToolTagSuffix checks if the string ends with a prefix of a known
// tool_call tag (e.g. "<tool", "<func", "<|tool"). Returns the number of
// trailing bytes that should be held back until the next chunk arrives.
// Returns 0 if no partial match is found.
func partialToolTagSuffix(s string) int {
	tags := []string{"<tool_call>", "<|tool_call|>", "<function=", "<parameter=", "<think>", "</think>"}
	maxHold := 0
	for _, tag := range tags {
		// Check every possible prefix of this tag (length 1..len(tag)-1)
		for prefixLen := len(tag) - 1; prefixLen >= 1; prefixLen-- {
			prefix := tag[:prefixLen]
			if strings.HasSuffix(s, prefix) && prefixLen > maxHold {
				maxHold = prefixLen
				break // found longest prefix match for this tag
			}
		}
	}
	return maxHold
}

// findToolCallTagStart returns the byte index of the first tool_call tag
// start in the given content, or -1 if none found. Used by the streaming
// loop to suppress tool_call output from being displayed to the user.
func findToolCallTagStart(s string) int {
	patterns := []string{"<tool_call>", "<|tool_call|>", "<function="}
	minIdx := -1
	for _, p := range patterns {
		if idx := strings.Index(s, p); idx >= 0 && (minIdx < 0 || idx < minIdx) {
			minIdx = idx
		}
	}
	return minIdx
}

// StripToolCallTags removes tool_call/function tags and their content from
// a string, returning only the human-readable text. Used to clean the
// stream buffer when text-based tool calls were detected.
func StripToolCallTags(s string) string {
	s = stripThinkTags(s)

	// Strip <tool_call>...</tool_call>
	for {
		start := strings.Index(s, "<tool_call>")
		if start == -1 {
			break
		}
		end := strings.Index(s[start:], "</tool_call>")
		if end == -1 {
			s = s[:start]
			break
		}
		s = s[:start] + s[start+end+len("</tool_call>"):]
	}
	// Strip <|tool_call|>...<|/tool_call|>
	for {
		start := strings.Index(s, "<|tool_call|>")
		if start == -1 {
			break
		}
		end := strings.Index(s[start:], "<|/tool_call|>")
		if end == -1 {
			s = s[:start]
			break
		}
		s = s[:start] + s[start+end+len("<|/tool_call|>"):]
	}
	// Strip <function=...>...</function> and <function=...>...</tool_call>
	for {
		start := strings.Index(s, "<function=")
		if start == -1 {
			break
		}
		// Try </function> first
		end := strings.Index(s[start:], "</function>")
		closeLen := len("</function>")
		if end == -1 {
			end = strings.Index(s[start:], "</tool_call>")
			closeLen = len("</tool_call>")
		}
		if end == -1 {
			s = s[:start]
			break
		}
		s = s[:start] + s[start+end+closeLen:]
	}

	return strings.TrimSpace(s)
}

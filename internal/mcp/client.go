package mcp

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os/exec"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/kimjiwon/tgc/internal/config"
)

// Client is an MCP client for a single server.
type Client struct {
	server MCPServer

	// stdio transport
	cmd    *exec.Cmd
	stdin  io.WriteCloser
	stdout *bufio.Scanner

	// sse transport
	httpClient *http.Client
	sseURL     string

	mu      sync.Mutex
	nextID  atomic.Int64
	pending map[int64]chan JSONRPCResponse

	tools []Tool
}

// NewClient creates a new MCP client for the given server config.
func NewClient(server MCPServer) *Client {
	return &Client{
		server:  server,
		pending: make(map[int64]chan JSONRPCResponse),
	}
}

// MCPServer is re-exported here for internal use (avoids import cycle).
// The canonical type lives in config. This alias is defined in manager.go.

// Start initializes the transport connection and performs the MCP handshake.
func (c *Client) Start() error {
	config.DebugLog("[MCP] starting client name=%s transport=%s", c.server.Name, c.server.Transport)
	switch c.server.Transport {
	case "stdio":
		return c.startStdio()
	case "sse":
		return c.startSSE()
	default:
		return fmt.Errorf("unknown transport: %s", c.server.Transport)
	}
}

func (c *Client) startStdio() error {
	if c.server.Command == "" {
		return fmt.Errorf("stdio transport requires command")
	}
	cmd := exec.Command(c.server.Command, c.server.Args...)
	for k, v := range c.server.Env {
		cmd.Env = append(cmd.Env, k+"="+v)
	}

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return fmt.Errorf("stdin pipe: %w", err)
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("stdout pipe: %w", err)
	}
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("start command: %w", err)
	}

	c.cmd = cmd
	c.stdin = stdin
	c.stdout = bufio.NewScanner(stdout)

	go c.readLoop()

	if err := c.Initialize(); err != nil {
		return fmt.Errorf("initialize: %w", err)
	}
	return nil
}

func (c *Client) startSSE() error {
	if c.server.URL == "" {
		return fmt.Errorf("sse transport requires url")
	}
	c.httpClient = &http.Client{}
	c.sseURL = c.server.URL

	if err := c.Initialize(); err != nil {
		return fmt.Errorf("initialize: %w", err)
	}
	return nil
}

// readLoop continuously reads JSON-RPC responses from stdio stdout.
func (c *Client) readLoop() {
	for c.stdout.Scan() {
		line := c.stdout.Text()
		if line == "" {
			continue
		}
		config.DebugLog("[MCP-STDIO] recv name=%s line=%s", c.server.Name, truncate(line, 200))
		var resp JSONRPCResponse
		if err := json.Unmarshal([]byte(line), &resp); err != nil {
			config.DebugLog("[MCP-STDIO] parse error: %v", err)
			continue
		}
		c.mu.Lock()
		ch, ok := c.pending[int64(resp.ID)]
		c.mu.Unlock()
		if ok {
			ch <- resp
		}
	}
}

// sendRequest sends a JSON-RPC request and waits for the response.
func (c *Client) sendRequest(method string, params interface{}) (JSONRPCResponse, error) {
	id := c.nextID.Add(1)
	req := JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      int(id),
		Method:  method,
		Params:  params,
	}

	data, err := json.Marshal(req)
	if err != nil {
		return JSONRPCResponse{}, fmt.Errorf("marshal request: %w", err)
	}

	config.DebugLog("[MCP] send name=%s method=%s id=%d", c.server.Name, method, id)

	switch c.server.Transport {
	case "stdio":
		return c.sendStdio(id, data)
	case "sse":
		return c.sendSSE(data)
	default:
		return JSONRPCResponse{}, fmt.Errorf("unknown transport: %s", c.server.Transport)
	}
}

func (c *Client) sendStdio(id int64, data []byte) (JSONRPCResponse, error) {
	ch := make(chan JSONRPCResponse, 1)
	c.mu.Lock()
	c.pending[id] = ch
	c.mu.Unlock()

	defer func() {
		c.mu.Lock()
		delete(c.pending, id)
		c.mu.Unlock()
	}()

	c.mu.Lock()
	_, err := fmt.Fprintf(c.stdin, "%s\n", data)
	c.mu.Unlock()
	if err != nil {
		return JSONRPCResponse{}, fmt.Errorf("write stdin: %w", err)
	}

	resp := <-ch
	return resp, nil
}

func (c *Client) sendSSE(data []byte) (JSONRPCResponse, error) {
	req, err := http.NewRequest("POST", c.sseURL, strings.NewReader(string(data)))
	if err != nil {
		return JSONRPCResponse{}, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json, text/event-stream")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return JSONRPCResponse{}, fmt.Errorf("http request: %w", err)
	}
	defer resp.Body.Close()

	contentType := resp.Header.Get("Content-Type")
	if strings.Contains(contentType, "text/event-stream") {
		return c.parseSSEResponse(resp.Body)
	}

	var rpcResp JSONRPCResponse
	if err := json.NewDecoder(resp.Body).Decode(&rpcResp); err != nil {
		return JSONRPCResponse{}, fmt.Errorf("decode response: %w", err)
	}
	return rpcResp, nil
}

// parseSSEResponse reads SSE events and extracts the first data JSON-RPC response.
func (c *Client) parseSSEResponse(body io.Reader) (JSONRPCResponse, error) {
	scanner := bufio.NewScanner(body)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "data: ") {
			payload := strings.TrimPrefix(line, "data: ")
			var resp JSONRPCResponse
			if err := json.Unmarshal([]byte(payload), &resp); err != nil {
				continue
			}
			return resp, nil
		}
	}
	return JSONRPCResponse{}, fmt.Errorf("no data event received")
}

// Initialize sends the MCP initialize handshake.
func (c *Client) Initialize() error {
	params := InitializeParams{
		ProtocolVersion: "2024-11-05",
		Capabilities:    ClientCaps{},
		ClientInfo:      ClientInfo{Name: "techai", Version: "1.0"},
	}
	resp, err := c.sendRequest("initialize", params)
	if err != nil {
		return err
	}
	if resp.Error != nil {
		return fmt.Errorf("initialize error %d: %s", resp.Error.Code, resp.Error.Message)
	}

	// Send initialized notification (no response expected)
	notif := map[string]interface{}{
		"jsonrpc": "2.0",
		"method":  "notifications/initialized",
	}
	if c.server.Transport == "stdio" {
		data, _ := json.Marshal(notif)
		c.mu.Lock()
		fmt.Fprintf(c.stdin, "%s\n", data)
		c.mu.Unlock()
	}

	config.DebugLog("[MCP] initialized name=%s", c.server.Name)
	return nil
}

// ListTools fetches the tool list from the MCP server.
func (c *Client) ListTools() ([]Tool, error) {
	resp, err := c.sendRequest("tools/list", nil)
	if err != nil {
		return nil, err
	}
	if resp.Error != nil {
		return nil, fmt.Errorf("tools/list error %d: %s", resp.Error.Code, resp.Error.Message)
	}

	data, err := json.Marshal(resp.Result)
	if err != nil {
		return nil, fmt.Errorf("marshal result: %w", err)
	}
	var result ListToolsResult
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("unmarshal tools: %w", err)
	}

	c.tools = result.Tools
	config.DebugLog("[MCP] listed tools name=%s count=%d", c.server.Name, len(c.tools))
	return result.Tools, nil
}

// CallTool executes a tool on the MCP server.
func (c *Client) CallTool(name string, args map[string]interface{}) (string, error) {
	params := CallToolParams{Name: name, Arguments: args}
	resp, err := c.sendRequest("tools/call", params)
	if err != nil {
		return "", err
	}
	if resp.Error != nil {
		return "", fmt.Errorf("tools/call error %d: %s", resp.Error.Code, resp.Error.Message)
	}

	data, err := json.Marshal(resp.Result)
	if err != nil {
		return "", fmt.Errorf("marshal result: %w", err)
	}
	var result CallToolResult
	if err := json.Unmarshal(data, &result); err != nil {
		return "", fmt.Errorf("unmarshal call result: %w", err)
	}

	var parts []string
	for _, c := range result.Content {
		if c.Text != "" {
			parts = append(parts, c.Text)
		}
	}
	out := strings.Join(parts, "\n")
	if result.IsError {
		return "", fmt.Errorf("tool error: %s", out)
	}
	return out, nil
}

// Stop shuts down the client.
func (c *Client) Stop() {
	if c.stdin != nil {
		c.stdin.Close()
	}
	if c.cmd != nil {
		c.cmd.Wait()
	}
	config.DebugLog("[MCP] stopped name=%s", c.server.Name)
}

// Tools returns the cached tool list (populated after ListTools).
func (c *Client) Tools() []Tool {
	return c.tools
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}

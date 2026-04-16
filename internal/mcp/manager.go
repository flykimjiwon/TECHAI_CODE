package mcp

import (
	"fmt"
	"strings"

	"github.com/kimjiwon/tgc/internal/config"
)

// MCPServer mirrors config.MCPServer to avoid import cycles.
// The Manager is initialized from config.MCPServer values.
type MCPServer = config.MCPServer

// Manager holds multiple MCP clients, one per configured server.
type Manager struct {
	servers []MCPServer
	clients []*Client
}

// NewManager creates a Manager for the given server configs.
func NewManager(servers []MCPServer) *Manager {
	return &Manager{servers: servers}
}

// Start initializes all clients and collects their tools.
func (m *Manager) Start() error {
	for _, srv := range m.servers {
		c := NewClient(srv)
		if err := c.Start(); err != nil {
			config.DebugLog("[MCP-MGR] failed to start server=%s: %v", srv.Name, err)
			// Continue to next server rather than aborting all
			m.clients = append(m.clients, nil)
			continue
		}
		if _, err := c.ListTools(); err != nil {
			config.DebugLog("[MCP-MGR] failed to list tools server=%s: %v", srv.Name, err)
		}
		m.clients = append(m.clients, c)
		config.DebugLog("[MCP-MGR] started server=%s tools=%d", srv.Name, len(c.Tools()))
	}
	return nil
}

// Stop shuts down all running clients.
func (m *Manager) Stop() {
	for _, c := range m.clients {
		if c != nil {
			c.Stop()
		}
	}
}

// AllTools returns the combined tool list from all connected servers,
// with names prefixed as mcp_{servername}_.
func (m *Manager) AllTools() []MCPTool {
	var out []MCPTool
	for i, c := range m.clients {
		if c == nil {
			continue
		}
		srv := m.servers[i]
		for _, t := range c.Tools() {
			out = append(out, MCPTool{
				Name:        mcpToolName(srv.Name, t.Name),
				Description: t.Description,
				InputSchema: t.InputSchema,
				ServerName:  srv.Name,
				OrigName:    t.Name,
			})
		}
	}
	return out
}

// MCPTool is a tool from an MCP server with routing metadata.
type MCPTool struct {
	Name        string
	Description string
	InputSchema interface{}
	ServerName  string
	OrigName    string
}

// CallTool routes a tool call to the correct server by prefixed name.
func (m *Manager) CallTool(name string, args map[string]interface{}) (string, error) {
	for i, c := range m.clients {
		if c == nil {
			continue
		}
		srv := m.servers[i]
		for _, t := range c.Tools() {
			if mcpToolName(srv.Name, t.Name) == name {
				config.DebugLog("[MCP-MGR] call server=%s tool=%s", srv.Name, t.Name)
				return c.CallTool(t.Name, args)
			}
		}
	}
	return "", fmt.Errorf("no MCP server has tool: %s", name)
}

// Status returns a human-readable status string for the /mcp command.
func (m *Manager) Status() string {
	if len(m.servers) == 0 {
		return "MCP 서버 설정 없음 (config.yaml의 mcp.servers 확인)"
	}
	var lines []string
	for i, srv := range m.servers {
		c := m.clients[i]
		if c == nil {
			lines = append(lines, fmt.Sprintf("  ✗ %s (%s) — 연결 실패", srv.Name, srv.Transport))
			continue
		}
		lines = append(lines, fmt.Sprintf("  ✓ %s (%s) — %d 도구", srv.Name, srv.Transport, len(c.Tools())))
		for _, t := range c.Tools() {
			lines = append(lines, fmt.Sprintf("      • %s", mcpToolName(srv.Name, t.Name)))
		}
	}
	return strings.Join(lines, "\n")
}

func mcpToolName(serverName, toolName string) string {
	safe := strings.ReplaceAll(serverName, "-", "_")
	safe = strings.ReplaceAll(safe, " ", "_")
	return "mcp_" + safe + "_" + toolName
}

package app

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"charm.land/bubbles/v2/textarea"
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	openai "github.com/sashabaranov/go-openai"

	tgc "github.com/kimjiwon/tgc"
	"github.com/kimjiwon/tgc/internal/config"
	"github.com/kimjiwon/tgc/internal/knowledge"
	"github.com/kimjiwon/tgc/internal/llm"
	"github.com/kimjiwon/tgc/internal/tools"
	"github.com/kimjiwon/tgc/internal/ui"
)

type streamChunkMsg struct {
	content   string
	done      bool
	err       error
	toolCalls []llm.ToolCallInfo
}

type toolResultMsg struct {
	results []toolResult
}

type toolResult struct {
	callID string
	name   string
	output string
}

type Model struct {
	cfg       config.Config
	client    *llm.Client
	activeTab int
	cwd       string // current working directory (abbreviated)

	// Single shared conversation (not per-mode)
	history []openai.ChatCompletionMessage
	msgs    []ui.Message

	projectCtx string // .techai.md content

	textarea textarea.Model
	viewport viewport.Model

	streaming    bool
	streamBuf    string
	streamCh     <-chan llm.StreamChunk
	streamCancel context.CancelFunc
	streamStart  time.Time
	lastChunkAt  time.Time
	lastElapsed  time.Duration
	tokenCount   int
	toolIter     int // tool loop iteration counter (max 20)
	pendingQueue []string // messages queued while streaming
	knowledgeInj *knowledge.Injector

	inSetup    bool
	setupInput textarea.Model
	setupCfg   config.Config

	width  int
	height int
	ready  bool
}

func NewModel(cfg config.Config, initialMode int, needsSetup bool) Model {
	ta := textarea.New()
	ta.Placeholder = ""
	ta.CharLimit = 4096
	ta.SetWidth(80)
	ta.SetHeight(1)
	ta.ShowLineNumbers = false
	// Remove background color from textarea
	styles := ta.Styles()
	styles.Focused.CursorLine = lipgloss.NewStyle()
	styles.Focused.EndOfBuffer = lipgloss.NewStyle()
	styles.Blurred.CursorLine = lipgloss.NewStyle()
	styles.Blurred.EndOfBuffer = lipgloss.NewStyle()
	ta.SetStyles(styles)
	ta.Focus()

	setupTa := textarea.New()
	setupTa.Placeholder = "tg_..."
	setupTa.CharLimit = 512
	setupTa.SetWidth(60)
	setupTa.SetHeight(1)
	setupTa.ShowLineNumbers = false

	vp := viewport.New(viewport.WithWidth(80), viewport.WithHeight(20))

	// Get abbreviated cwd
	cwd, _ := os.Getwd()
	cwdShort := filepath.Base(cwd)

	// Load project context
	projectCtx := ""
	if data, err := os.ReadFile(".techai.md"); err == nil && len(data) > 0 {
		projectCtx = "\n\n## Project Context (.techai.md)\n" + string(data)
	}
	projectCtx += llm.GatherSystemContext()

	// Initialize knowledge store
	var knowledgeInj *knowledge.Injector
	if knowledgeStore, err := knowledge.NewStore(tgc.KnowledgeFS); err == nil {
		knowledgeInj = knowledge.NewInjector(knowledgeStore, 8192)
		config.DebugLog("[KNOWLEDGE] loaded %d documents", knowledgeStore.DocCount())
	} else {
		config.DebugLog("[KNOWLEDGE] failed to load: %v", err)
	}

	m := Model{
		cfg:          cfg,
		activeTab:    initialMode,
		cwd:          cwdShort,
		projectCtx:   projectCtx,
		knowledgeInj: knowledgeInj,
		textarea:     ta,
		viewport:     vp,
		inSetup:      needsSetup,
		setupCfg:     config.DefaultConfig(),
		setupInput:   setupTa,
	}

	if needsSetup {
		m.setupInput.Focus()
	} else {
		m.client = llm.NewClient(cfg.API.BaseURL, cfg.API.APIKey)
	}

	// Single conversation with initial mode's system prompt
	mode := llm.Mode(initialMode)
	sysPrompt := llm.SystemPrompt(mode) + projectCtx
	m.history = []openai.ChatCompletionMessage{
		{Role: openai.ChatMessageRoleSystem, Content: sysPrompt},
	}
	m.msgs = []ui.Message{
		{Role: ui.RoleSystem, Content: ui.RenderLogo(), Timestamp: time.Now()},
		{Role: ui.RoleSystem, Content: ui.ModeInfoBox(initialMode, m.currentModel()), Timestamp: time.Now(), Tag: "modebox"},
	}

	if config.IsDebug() {
		m.msgs = append(m.msgs, ui.Message{
			Role:      ui.RoleSystem,
			Content:   fmt.Sprintf("[DEBUG MODE] 로그 파일: %s", config.DebugLogPath()),
			Timestamp: time.Now(),
		})
	}

	return m
}

func (m Model) Init() tea.Cmd {
	return textarea.Blink
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.inSetup {
		return m.updateSetup(msg)
	}

	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		if m.streaming {
			switch msg.String() {
			case "ctrl+c", "esc":
				m.cancelStream()
				return m, nil
			case "enter":
				// Queue message while streaming
				input := strings.TrimSpace(m.textarea.Value())
				if input != "" {
					m.pendingQueue = append(m.pendingQueue, input)
					m.textarea.Reset()
					m.textarea.SetHeight(1)
					m.recalcLayout()
					// Show queued indicator
					m.msgs = append(m.msgs, ui.Message{
						Role: ui.RoleUser, Content: input + " [대기중]", Timestamp: time.Now(),
					})
					m.updateViewport()
				}
				return m, nil
			case "shift+enter":
				m.textarea.InsertString("\n")
				return m, nil
			case "tab":
				return m, nil // ignore tab during streaming
			}
			// Forward other keys to textarea for typing
			var taCmd tea.Cmd
			m.textarea, taCmd = m.textarea.Update(msg)
			lines := strings.Count(m.textarea.Value(), "\n") + 1
			if lines > m.textarea.Height() && lines <= 10 {
				m.textarea.SetHeight(lines)
				m.recalcLayout()
			}
			return m, taCmd
		}

		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit

		case "ctrl+l":
			// Keep system prompt, clear conversation
			m.history = m.history[:1]
			m.msgs = m.msgs[:0]
			m.msgs = append(m.msgs,
				ui.Message{Role: ui.RoleSystem, Content: ui.RenderLogo(), Timestamp: time.Now()},
				ui.Message{Role: ui.RoleSystem, Content: ui.ModeInfoBox(m.activeTab, m.currentModel()), Timestamp: time.Now(), Tag: "modebox"},
			)
			m.streamBuf = ""
			m.tokenCount = 0
			m.lastElapsed = 0
			m.updateViewport()
			return m, nil

		case "tab":
			m.activeTab = (m.activeTab + 1) % llm.ModeCount
			// Update system prompt for new mode
			mode := llm.Mode(m.activeTab)
			m.history[0] = openai.ChatCompletionMessage{
				Role:    openai.ChatMessageRoleSystem,
				Content: llm.SystemPrompt(mode) + m.projectCtx,
			}
			// Remove previous mode info boxes, keep only the new one
			filtered := m.msgs[:0]
			for _, msg := range m.msgs {
				if msg.Tag != "modebox" {
					filtered = append(filtered, msg)
				}
			}
			m.msgs = append(filtered, ui.Message{
				Role:      ui.RoleSystem,
				Content:   ui.ModeInfoBox(m.activeTab, m.currentModel()),
				Timestamp: time.Now(),
				Tag:       "modebox",
			})
			m.updateViewport()
			return m, nil

		case "shift+enter":
			// Shift+Enter = newline
			m.textarea.InsertString("\n")
			lines := strings.Count(m.textarea.Value(), "\n") + 1
			if lines > m.textarea.Height() && lines <= 10 {
				m.textarea.SetHeight(lines)
				m.recalcLayout()
			}
			return m, nil

		case "enter":
			// Enter = send message
			input := strings.TrimSpace(m.textarea.Value())
			if input != "" {
				m.textarea.Reset()
				m.textarea.SetHeight(1)
				m.recalcLayout()
				if handled, cmd := m.handleSlashCommand(input); handled {
					return m, cmd
				}
				return m, m.sendMessage(input)
			}
			return m, nil

		case "pgup", "pgdown":
			var vpCmd tea.Cmd
			m.viewport, vpCmd = m.viewport.Update(msg)
			return m, vpCmd

		case "alt+up":
			m.viewport.ScrollUp(3)
			return m, nil
		case "alt+down":
			m.viewport.ScrollDown(3)
			return m, nil
		}

		// Default: forward to textarea
		var taCmd tea.Cmd
		m.textarea, taCmd = m.textarea.Update(msg)
		// Auto-grow/shrink textarea after content changes
		lines := strings.Count(m.textarea.Value(), "\n") + 1
		if lines > m.textarea.Height() && lines <= 10 {
			m.textarea.SetHeight(lines)
			m.recalcLayout()
		} else if lines < m.textarea.Height() {
			m.textarea.SetHeight(lines)
			m.recalcLayout()
		}
		return m, taCmd

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.ready = true
		m.recalcLayout()
		m.updateViewport()
		return m, nil

	case streamChunkMsg:
		if msg.err != nil {
			m.streaming = false
			m.streamCh = nil
			m.lastElapsed = time.Since(m.streamStart)
			config.DebugLog("[APP-STREAM] error after %v: %v", m.lastElapsed, msg.err)
			m.msgs = append(m.msgs, ui.Message{
				Role: ui.RoleSystem, Content: fmt.Sprintf("Error: %v", msg.err), Timestamp: time.Now(),
			})
			m.streamBuf = ""
			m.updateViewport()
			return m, nil
		}
		if msg.done {
			m.streamCh = nil

			// Check if AI wants to call tools
			if len(msg.toolCalls) > 0 {
				config.DebugLog("[APP-STREAM] done reason=tool_call | toolCalls=%d | bufLen=%d", len(msg.toolCalls), len(m.streamBuf))
				if m.streamBuf != "" {
					m.msgs = append(m.msgs, ui.Message{
						Role: ui.RoleAssistant, Content: m.streamBuf, Timestamp: time.Now(),
					})
				}

				assistantMsg := openai.ChatCompletionMessage{
					Role:    openai.ChatMessageRoleAssistant,
					Content: m.streamBuf,
				}
				var oaiToolCalls []openai.ToolCall
				for _, tc := range msg.toolCalls {
					oaiToolCalls = append(oaiToolCalls, openai.ToolCall{
						ID:   tc.ID,
						Type: openai.ToolTypeFunction,
						Function: openai.FunctionCall{
							Name:      tc.Name,
							Arguments: tc.Arguments,
						},
					})
				}
				assistantMsg.ToolCalls = oaiToolCalls
				m.history = append(m.history, assistantMsg)

				for _, tc := range msg.toolCalls {
					status := fmt.Sprintf(">> %s(%s)", tc.Name, truncateArgs(tc.Arguments, 60))
					m.msgs = append(m.msgs, ui.Message{
						Role: ui.RoleTool, Content: status, Timestamp: time.Now(),
					})
				}
				m.streamBuf = ""
				m.updateViewport()

				calls := msg.toolCalls
				return m, func() tea.Msg {
					var results []toolResult
					for _, tc := range calls {
						output := tools.Execute(tc.Name, tc.Arguments)
						results = append(results, toolResult{
							callID: tc.ID,
							name:   tc.Name,
							output: output,
						})
					}
					return toolResultMsg{results: results}
				}
			}

			// Normal completion (no tool calls)
			m.streaming = false
			m.lastElapsed = time.Since(m.streamStart)
			config.DebugLog("[APP-STREAM] done reason=normal | elapsed=%v | tokens=%d | bufLen=%d", m.lastElapsed, m.tokenCount, len(m.streamBuf))
			if m.streamBuf != "" {
				m.msgs = append(m.msgs, ui.Message{
					Role: ui.RoleAssistant, Content: m.streamBuf, Timestamp: time.Now(),
				})
				m.history = append(m.history, openai.ChatCompletionMessage{
					Role: openai.ChatMessageRoleAssistant, Content: m.streamBuf,
				})
			}
			m.streamBuf = ""
			m.updateViewport()

			// Auto-send next queued message
			if len(m.pendingQueue) > 0 {
				next := m.pendingQueue[0]
				m.pendingQueue = m.pendingQueue[1:]
				return m, m.sendMessage(next)
			}
			return m, nil
		}
		m.streamBuf += msg.content
		m.tokenCount++
		m.lastChunkAt = time.Now()
		m.updateViewport()
		return m, m.waitForNextChunk()

	case toolResultMsg:
		config.DebugLog("[APP-TOOL] received %d tool results | toolIter=%d/20", len(msg.results), m.toolIter+1)
		for _, r := range msg.results {
			config.DebugLog("[APP-TOOL] %s | resultLen=%d", r.name, len(r.output))
			m.history = append(m.history, openai.ChatCompletionMessage{
				Role:       openai.ChatMessageRoleTool,
				Content:    r.output,
				ToolCallID: r.callID,
			})
			preview := truncateArgs(r.output, 100)
			m.msgs = append(m.msgs, ui.Message{
				Role: ui.RoleTool, Content: fmt.Sprintf("<< %s: %s", r.name, preview), Timestamp: time.Now(),
			})
		}
		m.streamBuf = ""
		m.toolIter++
		m.updateViewport()

		if m.toolIter >= 20 {
			config.DebugLog("[APP-TOOL] loop limit reached (20 iterations)")
			m.streaming = false
			m.lastElapsed = time.Since(m.streamStart)
			m.msgs = append(m.msgs, ui.Message{
				Role: ui.RoleSystem, Content: "[tool loop limit — 20 iterations]", Timestamp: time.Now(),
			})
			m.updateViewport()
			return m, nil
		}

		return m, m.continueAfterTools()

	// Forward mouse wheel events to viewport for touchpad scroll
	case tea.MouseWheelMsg:
		var vpCmd tea.Cmd
		m.viewport, vpCmd = m.viewport.Update(msg)
		return m, vpCmd
	}

	return m, nil
}

func (m Model) updateSetup(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "enter":
			input := strings.TrimSpace(m.setupInput.Value())
			if input != "" {
				m.setupCfg.API.APIKey = input
			}
			_ = config.Save(m.setupCfg)
			m.cfg = m.setupCfg
			m.client = llm.NewClient(m.cfg.API.BaseURL, m.cfg.API.APIKey)
			m.inSetup = false
			m.ready = true
			m.recalcLayout()
			m.textarea.Focus()
			m.updateViewport()
			return m, nil
		}
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.ready = true
		m.setupInput.SetWidth(m.width - 10)
		m.recalcLayout()
		return m, nil
	}

	var cmd tea.Cmd
	m.setupInput, cmd = m.setupInput.Update(msg)
	return m, cmd
}

func (m *Model) handleSlashCommand(input string) (bool, tea.Cmd) {
	switch input {
	case "/setup":
		m.inSetup = true
		m.setupCfg = m.cfg
		m.setupInput.Reset()
		m.setupInput.Placeholder = "tg_..."
		m.setupInput.Focus()
		return true, nil

	case "/clear":
		m.history = m.history[:1]
		m.msgs = m.msgs[:0]
		m.msgs = append(m.msgs,
			ui.Message{Role: ui.RoleSystem, Content: ui.RenderLogo(), Timestamp: time.Now()},
			ui.Message{Role: ui.RoleSystem, Content: ui.ModeInfoBox(m.activeTab, m.currentModel()), Timestamp: time.Now(), Tag: "modebox"},
		)
		m.streamBuf = ""
		m.tokenCount = 0
		m.lastElapsed = 0
		m.updateViewport()
		return true, nil

	case "/help":
		help := `  Enter — 전송    Shift+Enter — 줄바꿈    /clear — 대화삭제    Ctrl+C — 종료`
		m.msgs = append(m.msgs, ui.Message{
			Role: ui.RoleSystem, Content: help, Timestamp: time.Now(),
		})
		m.updateViewport()
		return true, nil
	}
	return false, nil
}

func (m Model) View() tea.View {
	var content string

	if m.inSetup {
		content = m.viewSetup()
	} else if !m.ready {
		content = "\n  로딩중..."
	} else {
		vpContent := m.viewport.View()
		// Constrain viewport output to allocated height
		contentLines := strings.Split(vpContent, "\n")
		if len(contentLines) > m.viewport.Height() {
			contentLines = contentLines[:m.viewport.Height()]
			vpContent = strings.Join(contentLines, "\n")
		}

		// Input box with gray border
		inputBox := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#9CA3AF")).
			Width(m.width - 4).
			Render(m.textarea.View())

		// Status bar below input — includes cwd and hints
		modelID := m.currentModel()
		displayModel := modelID
		if info, ok := llm.Models[modelID]; ok {
			displayModel = info.DisplayName
		}
		elapsed := m.lastElapsed
		if m.streaming {
			elapsed = time.Since(m.streamStart)
		}
		statusBar := ui.RenderStatusBar(displayModel, m.tokenCount, elapsed, m.activeTab, m.cwd, m.width, config.IsDebug(), len(tools.ToolsForMode(m.activeTab)))

		content = lipgloss.JoinVertical(lipgloss.Left, vpContent, inputBox, statusBar)
	}

	v := tea.NewView(content)
	v.AltScreen = true
	v.MouseMode = tea.MouseModeCellMotion
	return v
}

func (m Model) viewSetup() string {
	title := lipgloss.NewStyle().Foreground(ui.ColorPrimary).Bold(true)
	dim := lipgloss.NewStyle().Foreground(ui.ColorTextDim)
	hint := lipgloss.NewStyle().Foreground(ui.ColorMuted)
	step := lipgloss.NewStyle().Foreground(ui.ColorSuccess).Bold(true)

	var b strings.Builder
	b.WriteString("\n\n")
	b.WriteString(title.Render("  택가이코드 설정"))
	b.WriteString("\n")
	b.WriteString(dim.Render("  OpenAI-compatible API 연결"))
	b.WriteString("\n\n")

	b.WriteString(step.Render("  API Key") + "\n\n")
	b.WriteString("  " + m.setupInput.View())

	b.WriteString("\n\n")
	b.WriteString(hint.Render("  Enter 다음 · Ctrl+C 종료"))
	return b.String()
}

func (m *Model) recalcLayout() {
	if m.width == 0 || m.height == 0 {
		return
	}
	inputH := m.textarea.Height() + 2
	fixed := inputH + 1 // input box + status bar
	vpHeight := m.height - fixed
	if vpHeight < 3 {
		vpHeight = 3
	}
	m.viewport.SetWidth(m.width)
	m.viewport.SetHeight(vpHeight)
	m.textarea.SetWidth(m.width - 6)
}

var spinnerFrames = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}

func (m *Model) streamStatus() string {
	elapsed := time.Since(m.streamStart)
	frame := spinnerFrames[int(elapsed.Milliseconds()/150)%len(spinnerFrames)]

	if m.lastChunkAt.IsZero() {
		// No chunks received yet
		return fmt.Sprintf("%s 연결중... (%.1fs)", frame, elapsed.Seconds())
	}

	sinceLastChunk := time.Since(m.lastChunkAt)
	tps := float64(0)
	if elapsed.Seconds() > 0 {
		tps = float64(m.tokenCount) / elapsed.Seconds()
	}

	if sinceLastChunk > 15*time.Second {
		return fmt.Sprintf("%s 응답없음 (%.0fs 대기중 · %dtok)", frame, sinceLastChunk.Seconds(), m.tokenCount)
	}
	if sinceLastChunk > 5*time.Second {
		return fmt.Sprintf("%s 응답지연... (%.0fs · %dtok · %.1ftok/s)", frame, elapsed.Seconds(), m.tokenCount, tps)
	}
	return fmt.Sprintf("%s 수신중 (%.1fs · %dtok · %.1ftok/s)", frame, elapsed.Seconds(), m.tokenCount, tps)
}

func (m *Model) updateViewport() {
	var stream string
	if m.streaming {
		stream = m.streamStatus()
	}
	content := ui.RenderMessages(m.msgs, stream, m.viewport.Width())
	m.viewport.SetContent(content)
	m.viewport.GotoBottom()
}

func (m Model) currentModel() string {
	switch m.activeTab {
	case 1: // Dev mode → qwen3-coder-30b
		return m.cfg.Models.Dev
	default: // Super, Plan → gpt-oss-120b
		return m.cfg.Models.Super
	}
}

// startStream creates a new streaming request with tool definitions.
func (m *Model) startStream() tea.Cmd {
	ctx, cancel := context.WithCancel(context.Background())
	m.streamCancel = cancel
	model := m.currentModel()

	// Compact the in-memory history before copying so snipped/truncated
	// content persists across tool iterations (Phase 1-1). Stages 1+2 only
	// — stage 3 LLM summary requires a separate call and is deferred.
	m.history = llm.Compact(m.history)

	history := make([]openai.ChatCompletionMessage, len(m.history))
	copy(history, m.history)
	toolDefs := tools.ToolsForMode(m.activeTab)

	modeName := "super"
	if m.activeTab == 1 {
		modeName = "dev"
	} else if m.activeTab == 2 {
		modeName = "plan"
	}
	config.DebugLog("[APP-STREAM] start | mode=%s | historyMsgs=%d | tools=%d | toolIter=%d/20", modeName, len(history), len(toolDefs), m.toolIter)

	m.streamCh = m.client.StreamChat(ctx, model, history, toolDefs)
	return m.waitForNextChunk()
}

func (m *Model) sendMessage(input string) tea.Cmd {
	m.msgs = append(m.msgs, ui.Message{
		Role: ui.RoleUser, Content: input, Timestamp: time.Now(),
	})

	// Inject knowledge context into system prompt
	if m.knowledgeInj != nil {
		knowledgeCtx := m.knowledgeInj.Inject(m.activeTab, input)
		if knowledgeCtx != "" {
			mode := llm.Mode(m.activeTab)
			sysPrompt := llm.SystemPrompt(mode) + m.projectCtx + knowledgeCtx
			m.history[0] = openai.ChatCompletionMessage{
				Role: openai.ChatMessageRoleSystem, Content: sysPrompt,
			}
			config.DebugLog("[KNOWLEDGE] injected %d chars for query: %s",
				len(knowledgeCtx), truncate(input, 50))
		}
	}

	m.history = append(m.history, openai.ChatCompletionMessage{
		Role: openai.ChatMessageRoleUser, Content: input,
	})
	m.streaming = true
	m.streamBuf = ""
	m.tokenCount = 0
	m.toolIter = 0
	m.streamStart = time.Now()
	m.lastChunkAt = time.Time{}
	m.updateViewport()

	return m.startStream()
}

// continueAfterTools starts a new stream after tool results are added to history.
func (m *Model) continueAfterTools() tea.Cmd {
	m.streamBuf = ""
	return m.startStream()
}

func (m *Model) waitForNextChunk() tea.Cmd {
	ch := m.streamCh
	if ch == nil {
		return nil
	}
	return func() tea.Msg {
		chunk, ok := <-ch
		if !ok {
			return streamChunkMsg{done: true}
		}
		return streamChunkMsg{
			content:   chunk.Content,
			done:      chunk.Done,
			err:       chunk.Err,
			toolCalls: chunk.ToolCalls,
		}
	}
}

func (m *Model) cancelStream() {
	if m.streamCancel != nil {
		m.streamCancel()
		m.streamCancel = nil
	}
	m.streaming = false
	m.streamCh = nil
	m.lastElapsed = time.Since(m.streamStart)
	if m.streamBuf != "" {
		m.msgs = append(m.msgs, ui.Message{
			Role: ui.RoleAssistant, Content: m.streamBuf + "\n\n[중단됨]", Timestamp: time.Now(),
		})
		m.history = append(m.history, openai.ChatCompletionMessage{
			Role: openai.ChatMessageRoleAssistant, Content: m.streamBuf,
		})
	}
	m.streamBuf = ""
	m.updateViewport()
}

func truncate(s string, n int) string {
	runes := []rune(s)
	if len(runes) <= n {
		return s
	}
	return string(runes[:n]) + "..."
}

func truncateArgs(s string, max int) string {
	s = strings.ReplaceAll(s, "\n", " ")
	runes := []rune(s)
	if len(runes) > max {
		return string(runes[:max]) + "..."
	}
	return s
}

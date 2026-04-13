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
	"github.com/kimjiwon/tgc/internal/agents"
	"github.com/kimjiwon/tgc/internal/config"
	"github.com/kimjiwon/tgc/internal/gitinfo"
	"github.com/kimjiwon/tgc/internal/knowledge"
	"github.com/kimjiwon/tgc/internal/llm"
	"github.com/kimjiwon/tgc/internal/session"
	"github.com/kimjiwon/tgc/internal/tools"
	"github.com/kimjiwon/tgc/internal/ui"
)

type streamChunkMsg struct {
	content   string
	done      bool
	err       error
	toolCalls []llm.ToolCallInfo
	usage     *openai.Usage
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

	// Session persistence (nil-safe; a nil store means "do not persist").
	store            *session.Store
	currentSessionID int64
	titleSet         bool // true once the first user message renamed the session

	// Git info cache: refreshed on startup, before each stream, and
	// after tool results (since edits may have changed the working tree).
	gitInfo gitinfo.Info

	// Auto mode: AI completes tasks autonomously without user input.
	autoMode bool
	autoIter int

	// Command palette (Ctrl+K): fuzzy-search slash commands.
	showPalette     bool
	paletteQuery    string
	paletteSelected int
	paletteFiltered []ui.PaletteItem

	// Menu overlay (Esc when not streaming): quick-access actions.
	showMenu     bool
	menuSelected int

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
	// envCtx and userDocsTOC are appended after env probe runs (below)

	// Initialize knowledge store (built-in embedded docs)
	var knowledgeInj *knowledge.Injector
	if knowledgeStore, err := knowledge.NewStore(tgc.KnowledgeFS); err == nil {
		knowledgeInj = knowledge.NewInjector(knowledgeStore, 8192)
		config.DebugLog("[KNOWLEDGE] loaded %d embedded documents", knowledgeStore.DocCount())
	} else {
		config.DebugLog("[KNOWLEDGE] failed to load: %v", err)
	}

	// Scan user knowledge docs (.tgc/knowledge/ or ~/.tgc/knowledge/)
	userDocs := knowledge.ScanUserDocs()
	knowledge.GlobalIndex = userDocs
	if userDocs.Count() > 0 {
		config.DebugLog("[USERDOCS] indexed %d user documents from %s", userDocs.Count(), userDocs.Root())
	}

	// Probe environment: detect installed tools (node, python, go, etc.)
	envResults := llm.ProbeEnvironment()
	envCtx := llm.FormatEnvironmentContext(envResults)
	config.DebugLog("[ENV] probed %d tools", len(envResults))

	// Append environment + user docs context to projectCtx
	projectCtx += envCtx
	if userDocs.Count() > 0 {
		projectCtx += userDocs.TableOfContents()
	}

	// Open persistent session store. Failures degrade to in-memory only
	// (nil store); we surface the error as a system message so the user
	// knows history will not survive a restart.
	var sessionStore *session.Store
	var sessionOpenErr error
	if home, herr := os.UserHomeDir(); herr == nil {
		dbPath := filepath.Join(home, ".tgc", "sessions.db")
		store, err := session.Open(dbPath)
		if err == nil {
			sessionStore = store
			config.DebugLog("[SESSION] opened store: %s", dbPath)
		} else {
			sessionOpenErr = err
			config.DebugLog("[SESSION] open failed: %v", err)
		}
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
		store:        sessionStore,
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

	// Create the first session row so subsequent AppendMessage calls
	// have a valid parent. Fall back to in-memory only on failure.
	if m.store != nil {
		id, err := m.store.CreateSession("untitled", initialMode, m.currentModel())
		if err == nil {
			m.currentSessionID = id
			// Persist the system prompt so restored sessions start with
			// the exact same context we are working with right now.
			_ = m.store.AppendMessage(id, m.history[0])
			config.DebugLog("[SESSION] created id=%d", id)
		} else {
			config.DebugLog("[SESSION] create failed: %v", err)
		}
	}
	if sessionOpenErr != nil {
		m.msgs = append(m.msgs, ui.Message{
			Role:      ui.RoleSystem,
			Content:   fmt.Sprintf("[SESSION] 저장소 열기 실패, 이번 세션은 메모리에만 보존됩니다: %v", sessionOpenErr),
			Timestamp: time.Now(),
		})
	}

	// Snapshot the git working tree so the HUD can render branch/dirty
	// state on the very first frame. This is silent when the cwd is not
	// a git repository.
	m.gitInfo = gitinfo.Fetch(".")

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
		// Command palette (Ctrl+K overlay) — captures all keys while open.
		if m.showPalette {
			switch msg.String() {
			case "esc":
				m.showPalette = false
				m.paletteQuery = ""
				m.updateViewport()
				return m, nil
			case "enter":
				if len(m.paletteFiltered) > 0 && m.paletteSelected < len(m.paletteFiltered) {
					action := m.paletteFiltered[m.paletteSelected].Action
					m.showPalette = false
					m.paletteQuery = ""
					if handled, cmd := m.handleSlashCommand(action); handled {
						return m, cmd
					}
				}
				return m, nil
			case "up":
				if m.paletteSelected > 0 {
					m.paletteSelected--
				}
				return m, nil
			case "down":
				if m.paletteSelected < len(m.paletteFiltered)-1 {
					m.paletteSelected++
				}
				return m, nil
			case "backspace":
				if len(m.paletteQuery) > 0 {
					m.paletteQuery = m.paletteQuery[:len(m.paletteQuery)-1]
					m.paletteFiltered = ui.FuzzyFilter(m.paletteQuery, ui.PaletteItems)
					m.paletteSelected = 0
				}
				return m, nil
			default:
				if len(msg.String()) == 1 {
					m.paletteQuery += msg.String()
					m.paletteFiltered = ui.FuzzyFilter(m.paletteQuery, ui.PaletteItems)
					m.paletteSelected = 0
				}
				return m, nil
			}
		}

		// Menu overlay (Esc overlay) — captures all keys while open.
		if m.showMenu {
			switch msg.String() {
			case "esc":
				m.showMenu = false
				return m, nil
			case "enter":
				action := ui.MenuActionFromIndex(m.menuSelected)
				m.showMenu = false
				if action != "" {
					if handled, cmd := m.handleSlashCommand(action); handled {
						return m, cmd
					}
				}
				return m, nil
			case "up":
				if m.menuSelected > 0 {
					m.menuSelected--
				}
				return m, nil
			case "down":
				if m.menuSelected < len(ui.MainMenuItems)-1 {
					m.menuSelected++
				}
				return m, nil
			}
			return m, nil
		}

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

		case "esc":
			m.showMenu = true
			m.menuSelected = 0
			return m, nil

		case "ctrl+k":
			m.showPalette = true
			m.paletteQuery = ""
			m.paletteSelected = 0
			m.paletteFiltered = ui.FuzzyFilter("", ui.PaletteItems)
			return m, nil

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

			// Prefer real token count from provider usage; keep the
			// chunk-count fallback when the provider omits it.
			if msg.usage != nil {
				if n := msg.usage.TotalTokens; n > 0 {
					m.tokenCount = n
				} else if n := msg.usage.PromptTokens + msg.usage.CompletionTokens; n > 0 {
					m.tokenCount = n
				}
			}

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
				m.persistMessage(assistantMsg)

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
				assistantMsg := openai.ChatCompletionMessage{
					Role: openai.ChatMessageRoleAssistant, Content: m.streamBuf,
				}
				m.history = append(m.history, assistantMsg)
				m.persistMessage(assistantMsg)
			}
			m.streamBuf = ""
			m.updateViewport()

			// Auto mode OR Deep Agent mode: check markers and continue
			isAuto := m.autoMode || m.activeTab == 1 // Deep Agent is always autonomous
			if isAuto {
				lastContent := ""
				if len(m.msgs) > 0 {
					lastContent = m.msgs[len(m.msgs)-1].Content
				}
				complete, pause := agents.CheckAutoMarkers(lastContent)
				if complete {
					m.autoMode = false
					m.autoIter = 0
					label := "[AUTO]"
					if m.activeTab == 1 {
						label = "[DEEP]"
					}
					m.msgs = append(m.msgs, ui.Message{
						Role: ui.RoleSystem, Content: label + " 작업 완료", Timestamp: time.Now(),
					})
					m.updateViewport()
					return m, nil
				}
				if pause {
					m.msgs = append(m.msgs, ui.Message{
						Role: ui.RoleSystem, Content: "[AUTO] 일시정지 — 입력을 기다립니다", Timestamp: time.Now(),
					})
					m.updateViewport()
					return m, nil
				}
				m.autoIter++
				maxIter := agents.MaxAutoIterations
				if m.activeTab == 1 {
					maxIter = agents.MaxDeepIterations // Deep Agent: 100 iterations
				}
				if m.autoIter < maxIter {
					return m, m.sendMessage("continue")
				}
				m.autoMode = false
				m.autoIter = 0
				m.msgs = append(m.msgs, ui.Message{
					Role: ui.RoleSystem, Content: fmt.Sprintf("[AUTO] 최대 반복 도달 (%d회)", maxIter), Timestamp: time.Now(),
				})
				m.updateViewport()
				return m, nil
			}

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
			toolMsg := openai.ChatCompletionMessage{
				Role:       openai.ChatMessageRoleTool,
				Content:    r.output,
				ToolCallID: r.callID,
			}
			m.history = append(m.history, toolMsg)
			m.persistMessage(toolMsg)
			preview := truncateArgs(r.output, 100)
			m.msgs = append(m.msgs, ui.Message{
				Role: ui.RoleTool, Content: fmt.Sprintf("<< %s: %s", r.name, preview), Timestamp: time.Now(),
			})
		}
		m.streamBuf = ""
		m.toolIter++
		// File-editing tools likely changed the working tree — refresh
		// the git snapshot so the HUD dirty indicator stays accurate.
		m.gitInfo = gitinfo.Fetch(".")
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
	// Split "command arg..." so /session <id> works. input is already
	// trimmed by the caller.
	cmd := input
	arg := ""
	if idx := strings.IndexByte(input, ' '); idx >= 0 {
		cmd = input[:idx]
		arg = strings.TrimSpace(input[idx+1:])
	}

	switch cmd {
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

	case "/new":
		// Start a fresh persisted session (current one stays in the DB).
		m.history = m.history[:1]
		m.msgs = m.msgs[:0]
		m.msgs = append(m.msgs,
			ui.Message{Role: ui.RoleSystem, Content: ui.RenderLogo(), Timestamp: time.Now()},
			ui.Message{Role: ui.RoleSystem, Content: ui.ModeInfoBox(m.activeTab, m.currentModel()), Timestamp: time.Now(), Tag: "modebox"},
		)
		m.streamBuf = ""
		m.tokenCount = 0
		m.lastElapsed = 0
		m.titleSet = false
		if m.store != nil {
			id, err := m.store.CreateSession("untitled", m.activeTab, m.currentModel())
			if err == nil {
				m.currentSessionID = id
				_ = m.store.AppendMessage(id, m.history[0])
				m.msgs = append(m.msgs, ui.Message{
					Role: ui.RoleSystem, Content: fmt.Sprintf("[SESSION] new id=%d", id), Timestamp: time.Now(),
				})
			}
		}
		m.updateViewport()
		return true, nil

	case "/sessions":
		if m.store == nil {
			m.msgs = append(m.msgs, ui.Message{
				Role: ui.RoleSystem, Content: "[SESSION] 저장소가 열려 있지 않습니다.", Timestamp: time.Now(),
			})
			m.updateViewport()
			return true, nil
		}
		list, err := m.store.ListSessions(10)
		if err != nil {
			m.msgs = append(m.msgs, ui.Message{
				Role: ui.RoleSystem, Content: fmt.Sprintf("[SESSION] 목록 조회 실패: %v", err), Timestamp: time.Now(),
			})
			m.updateViewport()
			return true, nil
		}
		if len(list) == 0 {
			m.msgs = append(m.msgs, ui.Message{
				Role: ui.RoleSystem, Content: "[SESSION] 저장된 세션이 없습니다.", Timestamp: time.Now(),
			})
			m.updateViewport()
			return true, nil
		}
		var sb strings.Builder
		sb.WriteString("  최근 세션 (/session <id> 로 불러오기)\n")
		for _, s := range list {
			marker := "  "
			if s.ID == m.currentSessionID {
				marker = "* "
			}
			sb.WriteString(fmt.Sprintf("%s#%d  %s  [%s]  %s\n",
				marker, s.ID, truncate(s.Title, 40), s.Model,
				s.UpdatedAt.Format("01-02 15:04")))
		}
		m.msgs = append(m.msgs, ui.Message{
			Role: ui.RoleSystem, Content: sb.String(), Timestamp: time.Now(),
		})
		m.updateViewport()
		return true, nil

	case "/session":
		if m.store == nil {
			m.msgs = append(m.msgs, ui.Message{
				Role: ui.RoleSystem, Content: "[SESSION] 저장소가 열려 있지 않습니다.", Timestamp: time.Now(),
			})
			m.updateViewport()
			return true, nil
		}
		if arg == "" {
			m.msgs = append(m.msgs, ui.Message{
				Role: ui.RoleSystem, Content: "[SESSION] 사용법: /session <id> (먼저 /sessions 로 목록 확인)", Timestamp: time.Now(),
			})
			m.updateViewport()
			return true, nil
		}
		var id int64
		if _, err := fmt.Sscanf(arg, "%d", &id); err != nil || id <= 0 {
			m.msgs = append(m.msgs, ui.Message{
				Role: ui.RoleSystem, Content: fmt.Sprintf("[SESSION] 잘못된 id: %q", arg), Timestamp: time.Now(),
			})
			m.updateViewport()
			return true, nil
		}
		meta, loaded, err := m.store.LoadSession(id)
		if err != nil {
			m.msgs = append(m.msgs, ui.Message{
				Role: ui.RoleSystem, Content: fmt.Sprintf("[SESSION] 로드 실패: %v", err), Timestamp: time.Now(),
			})
			m.updateViewport()
			return true, nil
		}
		// Restore history and rebuild the visible message stream from
		// the persisted record.
		m.history = loaded
		m.currentSessionID = meta.ID
		m.titleSet = meta.Title != "untitled"
		m.activeTab = meta.Mode
		m.msgs = m.msgs[:0]
		m.msgs = append(m.msgs,
			ui.Message{Role: ui.RoleSystem, Content: ui.RenderLogo(), Timestamp: time.Now()},
			ui.Message{Role: ui.RoleSystem, Content: ui.ModeInfoBox(m.activeTab, m.currentModel()), Timestamp: time.Now(), Tag: "modebox"},
			ui.Message{Role: ui.RoleSystem, Content: fmt.Sprintf("[SESSION] #%d '%s' 복원됨 (%d 메시지)", meta.ID, meta.Title, len(loaded)), Timestamp: time.Now()},
		)
		// Skip the system prompt at index 0 when rebuilding the UI view.
		for _, h := range loaded {
			switch h.Role {
			case openai.ChatMessageRoleUser:
				m.msgs = append(m.msgs, ui.Message{Role: ui.RoleUser, Content: h.Content, Timestamp: time.Now()})
			case openai.ChatMessageRoleAssistant:
				if h.Content != "" {
					m.msgs = append(m.msgs, ui.Message{Role: ui.RoleAssistant, Content: h.Content, Timestamp: time.Now()})
				}
				for _, tc := range h.ToolCalls {
					status := fmt.Sprintf(">> %s(%s)", tc.Function.Name, truncateArgs(tc.Function.Arguments, 60))
					m.msgs = append(m.msgs, ui.Message{Role: ui.RoleTool, Content: status, Timestamp: time.Now()})
				}
			case openai.ChatMessageRoleTool:
				preview := truncateArgs(h.Content, 100)
				m.msgs = append(m.msgs, ui.Message{Role: ui.RoleTool, Content: "<< " + preview, Timestamp: time.Now()})
			}
		}
		m.streamBuf = ""
		m.tokenCount = 0
		m.lastElapsed = 0
		m.updateViewport()
		return true, nil

	case "/auto":
		m.autoMode = !m.autoMode
		label := "OFF"
		if m.autoMode {
			label = "ON"
		}
		m.msgs = append(m.msgs, ui.Message{
			Role: ui.RoleSystem, Content: fmt.Sprintf("[AUTO] 자율 모드 %s (최대 %d회)", label, agents.MaxAutoIterations), Timestamp: time.Now(),
		})
		m.updateViewport()
		return true, nil

	case "/diagnostics":
		cwd, _ := os.Getwd()
		result, err := tools.RunDiagnostics(cwd, arg)
		if err != nil {
			result = fmt.Sprintf("Error: %v", err)
		}
		m.msgs = append(m.msgs, ui.Message{
			Role: ui.RoleSystem, Content: result, Timestamp: time.Now(),
		})
		m.updateViewport()
		return true, nil

	case "/git":
		// Force a fresh fetch so the user sees current state, not the
		// cached snapshot from the previous turn.
		m.gitInfo = gitinfo.Fetch(".")
		m.msgs = append(m.msgs, ui.Message{
			Role: ui.RoleSystem, Content: m.gitInfo.Summary(), Timestamp: time.Now(),
		})
		m.updateViewport()
		return true, nil

	case "/help":
		help := `  Enter — 전송    Shift+Enter — 줄바꿈    Tab — 모드 전환
  Ctrl+K — 커맨드 팔레트    Esc — 메뉴    Ctrl+C — 종료
  /new — 새 세션    /sessions — 목록    /session <id> — 복원
  /auto — 자율 모드    /diagnostics — 코드 진단    /git — 저장소 상태
  /clear — 화면 정리    /setup — API 키`
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
		ctxWindow := llm.GetCapability(modelID).ContextWindow
		statusBar := ui.RenderStatusBar(displayModel, m.tokenCount, ctxWindow, elapsed, m.activeTab, m.cwd, m.width, config.IsDebug(), len(tools.ToolsForMode(m.activeTab)), m.gitInfo.Label())

		// Overlay palette or menu on top of the viewport when active.
		if m.showPalette {
			overlay := ui.RenderPalette(m.paletteFiltered, m.paletteSelected, m.paletteQuery, m.width)
			vpContent = lipgloss.Place(m.width, m.viewport.Height(), lipgloss.Center, lipgloss.Center, overlay)
		} else if m.showMenu {
			overlay := ui.RenderMenu(ui.MainMenuItems, m.menuSelected, m.width)
			vpContent = lipgloss.Place(m.width, m.viewport.Height(), lipgloss.Center, lipgloss.Center, overlay)
		}

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
	// All modes use the super model by default (matching hanimo).
	// The dev model (qwen3-coder-30b) is available via TGC_MODEL_DEV
	// or config.yaml override if needed.
	return m.cfg.Models.Super
}

// startStream creates a new streaming request with tool definitions.
func (m *Model) startStream() tea.Cmd {
	ctx, cancel := context.WithCancel(context.Background())
	m.streamCancel = cancel
	model := m.currentModel()

	// Refresh git snapshot before each stream so the HUD reflects any
	// changes the user (or the previous tool iteration) made between
	// turns. Fetch is bounded to 500ms and degrades silently.
	m.gitInfo = gitinfo.Fetch(".")

	// Compact the in-memory history before copying so snipped/truncated
	// content persists across tool iterations (Phase 1-1). Stages 1+2
	// always run; stage 3 (LLM summary) only kicks in when the real token
	// count exceeds 90% of the model's context window (Phase 1.5).
	m.history = llm.Compact(m.history)

	ctxWindow := llm.GetCapability(model).ContextWindow
	if ui.ShouldAutoCompact(ui.ContextPercent(m.tokenCount, ctxWindow)) {
		// Target 50% of the window after summarization so we have room
		// for the next exchange without immediately re-triggering.
		target := ctxWindow / 2
		config.DebugLog("[APP-STREAM] auto-compact | tokens=%d window=%d target=%d",
			m.tokenCount, ctxWindow, target)
		m.history = llm.CompactWithLLM(ctx, m.client, model, m.history, target)
		// tokenCount will be refreshed by the provider Usage on the next
		// Done chunk; we leave it unchanged here so the HUD still shows
		// the pre-compact value until the real number arrives.
	}

	history := make([]openai.ChatCompletionMessage, len(m.history))
	copy(history, m.history)

	// When auto mode or Deep Agent mode is active, append the
	// autonomous-agent suffix to the system prompt copy (not the
	// original) so it only affects this request.
	if (m.autoMode || m.activeTab == 1) && len(history) > 0 {
		history[0].Content += agents.AutoPromptSuffix
	}

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

	userMsg := openai.ChatCompletionMessage{
		Role: openai.ChatMessageRoleUser, Content: input,
	}
	m.history = append(m.history, userMsg)
	m.persistMessage(userMsg)
	// First user message becomes the session title (truncated) so the
	// /sessions list shows something meaningful instead of "untitled".
	if !m.titleSet && m.store != nil && m.currentSessionID > 0 {
		if err := m.store.UpdateSessionTitle(m.currentSessionID, truncate(input, 60)); err == nil {
			m.titleSet = true
		}
	}
	m.streaming = true
	m.streamBuf = ""
	m.tokenCount = 0
	m.toolIter = 0
	m.streamStart = time.Now()
	m.lastChunkAt = time.Time{}
	m.updateViewport()

	return m.startStream()
}

// persistMessage writes a single chat message to the session store, if
// one is open. It is deliberately silent on failure — persistence is a
// best-effort side-channel, never a blocker for the live conversation.
func (m *Model) persistMessage(msg openai.ChatCompletionMessage) {
	if m.store == nil || m.currentSessionID == 0 {
		return
	}
	if err := m.store.AppendMessage(m.currentSessionID, msg); err != nil {
		config.DebugLog("[SESSION] append failed: %v", err)
	}
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
			usage:     chunk.Usage,
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

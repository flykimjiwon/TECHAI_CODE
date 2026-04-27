package app

import (
	"context"
	"encoding/json"
	"fmt"
	"runtime"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"

	"charm.land/bubbles/v2/textarea"
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/atotto/clipboard"
	"github.com/charmbracelet/x/ansi"
	openai "github.com/sashabaranov/go-openai"

	tgc "github.com/kimjiwon/tgc"
	"github.com/kimjiwon/tgc/internal/agents"
	"github.com/kimjiwon/tgc/internal/companion"
	"github.com/kimjiwon/tgc/internal/config"
	"github.com/kimjiwon/tgc/internal/gitinfo"
	"github.com/kimjiwon/tgc/internal/hooks"
	"github.com/kimjiwon/tgc/internal/knowledge"
	"github.com/kimjiwon/tgc/internal/llm"
	"github.com/kimjiwon/tgc/internal/mcp"
	"github.com/kimjiwon/tgc/internal/multi"
	"github.com/kimjiwon/tgc/internal/session"
	"github.com/kimjiwon/tgc/internal/tools"
	"github.com/kimjiwon/tgc/internal/ui"
)

type streamChunkMsg struct {
	content    string
	isThinking bool // true = reasoning_content (thinking phase)
	done       bool
	err        error
	toolCalls  []llm.ToolCallInfo
	usage      *openai.Usage
}

type toolResultMsg struct {
	results []toolResult
}

type toolResult struct {
	callID string
	name   string
	output string
}

// multiProgressMsg carries real-time progress from the multi-agent orchestrator.
type multiProgressMsg struct {
	progress multi.AgentProgress
}

// multiResultMsg carries the final merged result from the multi-agent pipeline.
type multiResultMsg struct {
	result multi.MergedResult
}

// slashResultMsg carries async slash command output back to the UI.
type slashResultMsg struct {
	content string
}

// spinnerTickMsg triggers periodic UI re-renders so the spinner stays animated.
type spinnerTickMsg struct{}

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

	// Multi-agent state
	multiEnabled    bool                       // global on/off
	multiStrategy   multi.Strategy             // current strategy
	multiAuto       bool                       // auto-detect mode
	multiRunning    bool                       // true while orchestrator is active
	multiProgress   []multi.AgentProgress      // latest progress from each agent
	multiCancel     context.CancelFunc         // cancel the orchestrator
	multiProgressCh <-chan multi.AgentProgress  // progress channel from orchestrator
	multiResultCh   <-chan multi.MergedResult   // result channel from orchestrator

	// Command palette (Ctrl+K): fuzzy-search slash commands.
	showPalette     bool
	paletteQuery    string
	paletteSelected int
	paletteFiltered []ui.PaletteItem

	// Menu overlay (Esc when not streaming): quick-access actions.
	showMenu     bool
	menuSelected int

	// Session picker overlay (from menu or /sessions).
	showSessionPicker     bool
	sessionPickerItems    []session.SessionMeta
	sessionPickerSelected int

	width  int
	height int
	ready  bool

	// Companion: browser-based real-time dashboard.
	companionHub    *companion.Hub
	companionServer *companion.Server
	companionPort   int

	// MCP: Model Context Protocol client manager.
	mcpManager *mcp.Manager

	// Input history: arrow keys cycle through previous inputs.
	inputHistory []string
	historyIdx   int // -1 = not browsing; 0 = most recent
	historyDraft string // saved current input when entering history

	// Custom commands loaded from .tgc/commands/ and ~/.tgc/commands/.
	customCommands map[string]string

	// Stream warning and retry
	streamWarnShown  bool
	wasThinking      bool   // tracks thinking→content transition in stream
	pendingPrefetch  string // auto-prefetched file contents, injected once then cleared
	streamRetries   int

	// Text selection: in-app mouse drag selection for copy
	selecting    bool
	selStartX    int
	selStartY    int // content line (viewport offset + screen Y)
	selEndX      int
	selEndY      int

	// Paste hint: shown above input box, cleared on next Enter
	pasteHint string

	// Memory: persistent project/global facts injected into system prompt.
	memoryStore *tools.MemoryStore

	// Hooks: lifecycle event handlers (pre/post tool use, session start/stop).
	hookManager *hooks.Manager
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

	// Initialize knowledge store (built-in embedded docs, filtered by project packs)
	var knowledgeInj *knowledge.Injector
	packs := parseKnowledgePacks(projectCtx)
	if len(packs) > 0 {
		config.DebugLog("[KNOWLEDGE] packs from .techai.md: %v", packs)
	}
	if knowledgeStore, err := knowledge.NewStore(tgc.KnowledgeFS, packs...); err == nil {
		knowledgeInj = knowledge.NewInjector(knowledgeStore, 8192)
		config.DebugLog("[KNOWLEDGE] loaded %d embedded documents (packs=%v)", knowledgeStore.DocCount(), packs)
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

	// Detect project type and framework
	projInfo := tools.DetectProject(".")
	projCtxStr := tools.FormatProjectContext(projInfo)
	if projCtxStr != "" {
		projectCtx += "\n\n## Detected Project\n" + projCtxStr
	}

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

	// Load custom commands from .tgc/commands/ and ~/.tgc/commands/
	customCmds := tools.LoadCustomCommands()
	if len(customCmds) > 0 {
		for name := range customCmds {
			config.DebugLog("[APP] custom command registered: /%s", name)
		}
		// Append custom commands to palette
		for name := range customCmds {
			ui.PaletteItems = append(ui.PaletteItems, ui.PaletteItem{
				Label:       "/" + name,
				Description: "Custom command",
				Action:      "/" + name,
			})
		}
	}

	m := Model{
		cfg:            cfg,
		activeTab:      initialMode,
		cwd:            cwdShort,
		projectCtx:     projectCtx,
		knowledgeInj:   knowledgeInj,
		textarea:       ta,
		viewport:       vp,
		inSetup:        needsSetup,
		setupCfg:       config.DefaultConfig(),
		setupInput:     setupTa,
		store:          sessionStore,
		customCommands: customCmds,
	}

	// Initialize lifecycle hooks
	m.hookManager = hooks.NewManager()

	// Initialize multi-agent config from loaded settings
	m.multiEnabled = cfg.Multi.Enabled
	m.multiAuto = cfg.Multi.Strategy == "auto"
	switch cfg.Multi.Strategy {
	case "review":
		m.multiStrategy = multi.StrategyReview
	case "consensus":
		m.multiStrategy = multi.StrategyConsensus
	case "scan":
		m.multiStrategy = multi.StrategyScan
	default:
		m.multiStrategy = multi.StrategyReview
	}

	if needsSetup {
		m.setupInput.Focus()
	} else {
		m.client = llm.NewClient(cfg.API.BaseURL, cfg.API.APIKey)

		// Wire Level 3 LLM fallback for knowledge search.
		// When keyword + BM25 search both fail, ask the LLM to pick docs.
		if knowledgeInj != nil {
			client := m.client
			model := cfg.Models.Super
			knowledgeInj.SetLLMSelector(func(prompt string) (string, error) {
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()
				return client.Chat(ctx, model, []openai.ChatCompletionMessage{
					{Role: openai.ChatMessageRoleUser, Content: prompt},
				})
			})
		}
	}

	// Initialize memory store
	m.memoryStore = tools.NewMemoryStore()
	memoryCtx := m.memoryStore.ForContext()

	// Single conversation with initial mode's system prompt
	mode := llm.Mode(initialMode)
	sysPrompt := llm.SystemPrompt(mode) + projectCtx + memoryCtx
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
			Content:   fmt.Sprintf("[DEBUG MODE] Log: %s", config.DebugLogPath()),
			Timestamp: time.Now(),
		})
	}

	// Show /init hint when .techai.md doesn't exist
	if _, err := os.Stat(".techai.md"); os.IsNotExist(err) {
		m.msgs = append(m.msgs, ui.Message{
			Role:      ui.RoleSystem,
			Content:   "  New here? Run /init to scan your project.",
			Timestamp: time.Now(),
		})
	}

	// Show newline shortcut hint based on OS
	newlineHint := "  Newline: Ctrl+J  |  /help for all shortcuts"
	if runtime.GOOS == "windows" {
		newlineHint = "  Newline: Ctrl+J (Windows)  |  /help for all shortcuts"
	}
	m.msgs = append(m.msgs, ui.Message{
		Role: ui.RoleSystem, Content: newlineHint, Timestamp: time.Now(),
	})

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

	// Initialize MCP clients from config. Done in background; failures
	// are logged and surfaced via /mcp status command.
	if len(cfg.MCP.Servers) > 0 {
		mgr := mcp.NewManager(cfg.MCP.Servers)
		if err := mgr.Start(); err != nil {
			config.DebugLog("[MCP] manager start error: %v", err)
		}
		m.mcpManager = mgr
		tools.MCPManager = mgr
		tools.RegisterMCPTools(mgr.AllTools())
		config.DebugLog("[MCP] manager started servers=%d tools=%d", len(cfg.MCP.Servers), len(mgr.AllTools()))
	}

	return m
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(textarea.Blink, spinnerTick())
}

// spinnerTick sends a tick every 200ms to keep the spinner animated.
func spinnerTick() tea.Cmd {
	return tea.Tick(200*time.Millisecond, func(time.Time) tea.Msg {
		return spinnerTickMsg{}
	})
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

		// Session picker overlay — captures all keys while open.
		if m.showSessionPicker {
			switch msg.String() {
			case "esc":
				m.showSessionPicker = false
				m.updateViewport()
				return m, nil
			case "enter":
				if len(m.sessionPickerItems) > 0 && m.sessionPickerSelected < len(m.sessionPickerItems) {
					picked := m.sessionPickerItems[m.sessionPickerSelected]
					m.showSessionPicker = false
					if picked.ID == m.currentSessionID {
						m.msgs = append(m.msgs, ui.Message{
							Role: ui.RoleSystem, Content: fmt.Sprintf("[SESSION] 이미 현재 세션입니다 (#%d)", picked.ID), Timestamp: time.Now(),
						})
						m.updateViewport()
						return m, nil
					}
					// Load the selected session
					if handled, cmd := m.handleSlashCommand(fmt.Sprintf("/session %d", picked.ID)); handled {
						return m, cmd
					}
				}
				return m, nil
			case "up":
				if m.sessionPickerSelected > 0 {
					m.sessionPickerSelected--
				}
				return m, nil
			case "down":
				if m.sessionPickerSelected < len(m.sessionPickerItems)-1 {
					m.sessionPickerSelected++
				}
				return m, nil
			}
			return m, nil
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
				if m.multiRunning {
					m.cancelMulti()
				} else {
					m.cancelStream()
				}
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
			case "shift+enter", "ctrl+j", "ctrl+enter":
				m.textarea.InsertString("\n")
				lines := strings.Count(m.textarea.Value(), "\n") + 1
				newH := min(lines, 10)
				if newH != m.textarea.Height() {
					m.textarea.SetHeight(newH)
					m.recalcLayout()
				}
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

		case "ctrl+u":
			// Clear input field
			m.textarea.Reset()
			m.textarea.SetHeight(1)
			m.historyIdx = -1
			m.historyDraft = ""
			m.pasteHint = ""
			m.recalcLayout()
			return m, nil


		case "ctrl+y":
			// Quick copy: last AI response to clipboard
			var target string
			for i := len(m.msgs) - 1; i >= 0; i-- {
				if m.msgs[i].Role == ui.RoleAssistant {
					target = m.msgs[i].Content
					break
				}
			}
			if target == "" {
				m.msgs = append(m.msgs, ui.Message{Role: ui.RoleSystem, Content: "복사할 AI 응답이 없습니다.", Timestamp: time.Now()})
			} else {
				if err := clipboard.WriteAll(target); err != nil {
					m.msgs = append(m.msgs, ui.Message{Role: ui.RoleSystem, Content: fmt.Sprintf("클립보드 복사 실패: %v", err), Timestamp: time.Now()})
				} else {
					runes := []rune(target)
					preview := string(runes)
					if len(runes) > 60 {
						preview = string(runes[:60]) + "..."
					}
					m.msgs = append(m.msgs, ui.Message{Role: ui.RoleSystem, Content: fmt.Sprintf("📋 복사됨 (%d자): %s", len(runes), preview), Timestamp: time.Now()})
				}
			}
			m.updateViewport()
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
			if m.companionHub != nil {
				m.companionHub.Emit("mode_change", map[string]interface{}{"mode": m.activeTab, "model": m.currentModel()})
			}
			m.updateViewport()
			return m, nil

		case "shift+enter", "ctrl+j", "ctrl+enter":
			// Shift+Enter or Ctrl+J = newline (Ctrl+J fallback for Windows CMD/PowerShell)
			m.textarea.InsertString("\n")
			lines := strings.Count(m.textarea.Value(), "\n") + 1
			newH := min(lines, 10)
			if newH != m.textarea.Height() {
				m.textarea.SetHeight(newH)
				m.recalcLayout()
			}
			return m, nil

		case "enter":
			// Enter = send message
			input := strings.TrimSpace(m.textarea.Value())
			if input != "" {
				// Save to input history
				m.inputHistory = append([]string{input}, m.inputHistory...)
				if len(m.inputHistory) > 100 {
					m.inputHistory = m.inputHistory[:100]
				}
				m.historyIdx = -1
				m.historyDraft = ""
				m.pasteHint = ""

				m.textarea.Reset()
				m.textarea.SetHeight(1)
				m.recalcLayout()
				if handled, cmd := m.handleSlashCommand(input); handled {
					return m, cmd
				}
				return m, m.sendMessage(input)
			}
			return m, nil

		case "up":
			// Up arrow: history when single-line with history, else scroll viewport
			if m.textarea.Height() == 1 && len(m.inputHistory) > 0 {
				if m.historyIdx == -1 {
					m.historyDraft = m.textarea.Value()
					m.historyIdx = 0
				} else if m.historyIdx < len(m.inputHistory)-1 {
					m.historyIdx++
				}
				m.textarea.Reset()
				m.textarea.InsertString(m.inputHistory[m.historyIdx])
				return m, nil
			}
			if m.textarea.Value() == "" {
				m.viewport.ScrollUp(3)
				return m, nil
			}

		case "down":
			// Down arrow: history forward when browsing, else scroll viewport
			if m.historyIdx >= 0 {
				m.historyIdx--
				m.textarea.Reset()
				if m.historyIdx < 0 {
					m.textarea.InsertString(m.historyDraft)
				} else {
					m.textarea.InsertString(m.inputHistory[m.historyIdx])
				}
				return m, nil
			}
			if m.textarea.Value() == "" {
				m.viewport.ScrollDown(3)
				return m, nil
			}

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

	case tea.PasteMsg:
		// Bracketed paste: insert pasted text into textarea.
		text := msg.Content
		if text == "" {
			return m, nil
		}
		m.textarea.InsertString(text)
		charCount := len([]rune(text))
		lineCount := strings.Count(text, "\n") + 1
		lines := strings.Count(m.textarea.Value(), "\n") + 1

		// Auto-grow textarea (up to 10 lines)
		newH := min(lines, 10)
		if newH != m.textarea.Height() {
			m.textarea.SetHeight(newH)
		}

		// Always show paste info
		if lineCount > 1 {
			m.pasteHint = fmt.Sprintf("[Pasted %d lines, %d chars — Enter to send, Ctrl+U to clear]", lineCount, charCount)
		} else {
			m.pasteHint = fmt.Sprintf("[Pasted %d chars — Enter to send, Ctrl+U to clear]", charCount)
		}
		m.recalcLayout()
		return m, nil

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.ready = true
		m.recalcLayout()
		m.updateViewport()
		return m, nil

	case streamChunkMsg:
		// Remove transient indicators when AI starts responding
		filtered := m.msgs[:0]
		for _, msg := range m.msgs {
			if msg.Tag != "processing" && msg.Tag != "stream-warn" {
				filtered = append(filtered, msg)
			}
		}
		m.msgs = filtered

		if msg.err != nil {
			m.lastElapsed = time.Since(m.streamStart)
			config.DebugLog("[APP-STREAM] error after %v: %v (retries=%d)", m.lastElapsed, msg.err, m.streamRetries)

			// Auto-retry on timeout (max 3 times)
			if strings.Contains(msg.err.Error(), "timeout") && m.streamRetries < 3 {
				m.streamRetries++
				if m.streamCancel != nil {
					m.streamCancel()
				}
				m.streamCh = nil
				m.msgs = append(m.msgs, ui.Message{
					Role: ui.RoleSystem, Content: fmt.Sprintf("  Retrying... (attempt %d/3)", m.streamRetries), Timestamp: time.Now(), Tag: "stream-warn",
				})
				m.updateViewport()
				m.streamStart = time.Now()
				m.lastChunkAt = time.Time{}
				m.streamWarnShown = false
				return m, m.startStream()
			}

			m.streaming = false
			m.streamCh = nil
			m.streamRetries = 0
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

				// Clean streamBuf when text-based tool calls were parsed from content.
				// The streaming suppression in client.go prevents most tag leakage,
				// but edge cases (tag split across chunks) may leave partial tags.
				if len(msg.toolCalls) > 0 && strings.HasPrefix(msg.toolCalls[0].ID, "text-tc-") {
					m.streamBuf = llm.StripToolCallTags(m.streamBuf)
				}

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
					status := formatToolCallPreview(tc.Name, tc.Arguments)
					m.msgs = append(m.msgs, ui.Message{
						Role: ui.RoleTool, Content: status, Timestamp: time.Now(),
					})
				}
				if m.companionHub != nil {
					for _, tc := range msg.toolCalls {
						m.companionHub.Emit("tool_call_start", map[string]interface{}{"id": tc.ID, "name": tc.Name, "args": truncateArgs(tc.Arguments, 200)})
					}
				}
				m.streamBuf = ""
				m.updateViewport()

				calls := msg.toolCalls
				hookMgr := m.hookManager
				return m, func() tea.Msg {
					var results []toolResult
					for _, tc := range calls {
						// Pre-tool hook: abort if hook returns exit code 2
						if hookMgr != nil {
							if hookMgr.RunPreToolUse(tc.Name, tc.Arguments) == hooks.HookFailAbort {
								config.DebugLog("[HOOKS] pre_tool_use aborted %s", tc.Name)
								results = append(results, toolResult{
									callID: tc.ID,
									name:   tc.Name,
									output: "Aborted by pre_tool_use hook",
								})
								continue
							}
						}

						// Per-tool timeout: 30 seconds max
						ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
						done := make(chan string, 1)
						go func(name, args string) {
							done <- tools.Execute(name, args)
						}(tc.Name, tc.Arguments)

						var output string
						select {
						case output = <-done:
							// Tool completed normally
						case <-ctx.Done():
							output = fmt.Sprintf("Error: tool %s timed out after 30s", tc.Name)
							config.DebugLog("[TOOL-TIMEOUT] %s timed out", tc.Name)
						}
						cancel()

						// Post-tool hook
						if hookMgr != nil {
							hookMgr.RunPostToolUse(tc.Name, tc.Arguments, output)
						}

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
			if m.companionHub != nil {
				m.companionHub.Emit("stream_done", map[string]interface{}{"content": m.streamBuf, "tokens": m.tokenCount, "elapsed": m.lastElapsed.Seconds()})
			}
			config.DebugLog("[APP-STREAM] done reason=normal | elapsed=%v | tokens=%d | bufLen=%d", m.lastElapsed, m.tokenCount, len(m.streamBuf))

			// Empty response detection — likely context overflow. Auto-compact and retry.
			if m.streamBuf == "" && m.toolIter > 0 && m.streamRetries < 3 {
				m.streamRetries++
				config.DebugLog("[APP-EMPTY] empty response after %d tools, compacting and retrying (attempt %d)", m.toolIter, m.streamRetries)
				// Force compact history — more aggressive each retry
				model := m.currentModel()
				ctxWindow := llm.GetCapability(model).ContextWindow
				// Retry 1: 33% (window/3), Retry 2: 25% (window/4), Retry 3: 20% (window/5)
				divisor := 2 + m.streamRetries // 3, 4, 5
				target := ctxWindow / divisor
				config.DebugLog("[APP-EMPTY] compact target=%d (window/%d)", target, divisor)
				ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
				m.history = llm.CompactWithLLM(ctx, m.client, model, m.history, target)
				cancel()
				m.msgs = append(m.msgs, ui.Message{
					Role: ui.RoleSystem, Content: fmt.Sprintf("  Compacting context and retrying... (attempt %d/3)", m.streamRetries), Timestamp: time.Now(), Tag: "stream-warn",
				})
				m.updateViewport()
				m.streaming = true
				m.streamStart = time.Now()
				m.lastChunkAt = time.Time{}
				m.streamWarnShown = false
				return m, m.startStream()
			}

			// All retries exhausted — show failure message so user isn't stuck
			if m.streamBuf == "" && m.toolIter > 0 {
				config.DebugLog("[APP-EMPTY] all retries exhausted after %d tools", m.toolIter)
				m.msgs = append(m.msgs, ui.Message{
					Role: ui.RoleSystem, Content: "  빈 응답이 반복되었습니다. 다시 시도하거나 질문을 더 구체적으로 입력해주세요.", Timestamp: time.Now(), Tag: "stream-warn",
				})
				m.updateViewport()
			}

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
		// Track thinking → content transition
		if msg.isThinking && !m.wasThinking && m.streamBuf == "" {
			// First thinking chunk — add marker
			m.streamBuf += "💭 "
			m.wasThinking = true
		} else if !msg.isThinking && m.wasThinking {
			// Transition from thinking to actual content — add separator
			m.streamBuf += "\n\n---\n\n"
			m.wasThinking = false
		}
		m.streamBuf += msg.content
		if m.companionHub != nil {
			m.companionHub.Emit("stream_chunk", map[string]interface{}{"content": msg.content, "total": len(m.streamBuf)})
		}
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
			preview := formatToolResultPreview(r.name, r.output)
			m.msgs = append(m.msgs, ui.Message{
				Role: ui.RoleTool, Content: preview, Timestamp: time.Now(),
			})
		}
		if m.companionHub != nil {
			for _, r := range msg.results {
				m.companionHub.Emit("tool_result", map[string]interface{}{"name": r.name, "output": truncateArgs(r.output, 500)})
			}
		}
		m.streamBuf = ""
		m.toolIter++
		// File-editing tools likely changed the working tree — refresh
		// the git snapshot so the HUD dirty indicator stays accurate.
		m.gitInfo = gitinfo.Fetch(".")

		// Show processing indicator so user knows AI is working
		m.msgs = append(m.msgs, ui.Message{
			Role: ui.RoleSystem, Content: fmt.Sprintf("  🔧 도구 실행 완료 (%d/20) — 다음 단계 진행중...", m.toolIter), Timestamp: time.Now(), Tag: "processing",
		})
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

	case spinnerTickMsg:
		// Re-render to animate spinner during streaming/tool execution.
		if m.streaming {
			m.updateViewport()
		}
		// Always keep the tick chain alive. When not streaming, the tick
		// is cheap (no updateViewport) and ensures the spinner starts
		// immediately when the next stream begins — no restart needed.
		return m, spinnerTick()

	case multiProgressMsg:
		p := msg.progress
		// Update or append progress entry
		found := false
		for i, existing := range m.multiProgress {
			if existing.Agent == p.Agent {
				m.multiProgress[i] = p
				found = true
				break
			}
		}
		if !found {
			m.multiProgress = append(m.multiProgress, p)
		}
		if m.companionHub != nil {
			m.companionHub.Emit("multi_progress", map[string]interface{}{"agent": p.Agent.String(), "status": p.Status, "detail": p.Detail, "tokens": p.Tokens})
		}
		m.updateViewport()
		// Continue waiting for more progress
		return m, m.waitForNextMulti()

	case slashResultMsg:
		m.msgs = append(m.msgs, ui.Message{Role: ui.RoleSystem, Content: msg.content, Timestamp: time.Now()})
		m.updateViewport()
		return m, nil

	case multiResultMsg:
		m.multiRunning = false
		m.multiCancel = nil
		m.multiProgress = nil
		m.lastElapsed = time.Since(m.streamStart)
		m.streaming = false

		result := msg.result
		config.DebugLog("[APP-MULTI] result strategy=%s elapsed=%v tokens=%d",
			result.Strategy, result.Elapsed, result.TotalTokens)

		if result.FinalOutput != "" {
			m.msgs = append(m.msgs, ui.Message{
				Role: ui.RoleAssistant, Content: result.FinalOutput, Timestamp: time.Now(),
			})
			assistantMsg := openai.ChatCompletionMessage{
				Role: openai.ChatMessageRoleAssistant, Content: result.FinalOutput,
			}
			m.history = append(m.history, assistantMsg)
			m.persistMessage(assistantMsg)
		}

		// Show summary
		m.msgs = append(m.msgs, ui.Message{
			Role: ui.RoleSystem,
			Content: fmt.Sprintf("  [Multi 완료] %s | %dtok | %.1fs",
				result.Strategy, result.TotalTokens, result.Elapsed.Seconds()),
			Timestamp: time.Now(),
		})
		m.tokenCount = result.TotalTokens
		if m.companionHub != nil {
			m.companionHub.Emit("multi_result", map[string]interface{}{"strategy": result.Strategy.String(), "tokens": result.TotalTokens, "elapsed": result.Elapsed.Seconds(), "output": truncateArgs(result.FinalOutput, 500)})
		}
		m.updateViewport()
		return m, nil

	// Forward mouse wheel events to viewport for touchpad scroll
	case tea.MouseWheelMsg:
		m.clearSelection()
		var vpCmd tea.Cmd
		m.viewport, vpCmd = m.viewport.Update(msg)
		return m, vpCmd

	// In-app text selection: click to start, drag to select, release to copy
	case tea.MouseClickMsg:
		if msg.Button == tea.MouseLeft && msg.Y < m.viewport.Height() {
			m.selecting = true
			m.selStartX = msg.X
			m.selStartY = m.viewport.YOffset() + msg.Y
			m.selEndX = msg.X
			m.selEndY = m.selStartY
		}
		return m, nil

	case tea.MouseMotionMsg:
		if m.selecting && msg.Y < m.viewport.Height() {
			m.selEndX = msg.X
			m.selEndY = m.viewport.YOffset() + msg.Y
		}
		return m, nil

	case tea.MouseReleaseMsg:
		if m.selecting {
			m.selecting = false
			text := m.extractSelectedText()
			if text != "" {
				if err := clipboard.WriteAll(text); err == nil {
					runes := []rune(text)
					preview := string(runes)
					if len(runes) > 60 {
						preview = string(runes[:60]) + "..."
					}
					m.msgs = append(m.msgs, ui.Message{
						Role: ui.RoleSystem, Content: fmt.Sprintf("📋 선택 복사됨 (%d자): %s", len(runes), preview), Timestamp: time.Now(),
					})
				}
			}
			m.clearSelection()
			m.updateViewport()
		}
		return m, nil
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
		if m.companionHub != nil {
			m.companionHub.Emit("session_create", map[string]interface{}{"id": m.currentSessionID})
		}
		return true, nil

	case "/sessions":
		if m.store == nil {
			m.msgs = append(m.msgs, ui.Message{
				Role: ui.RoleSystem, Content: "[SESSION] 저장소가 열려 있지 않습니다.", Timestamp: time.Now(),
			})
			m.updateViewport()
			return true, nil
		}
		list, err := m.store.ListSessions(20)
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
		m.sessionPickerItems = list
		m.sessionPickerSelected = 0
		m.showSessionPicker = true
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
		if m.companionHub != nil {
			m.companionHub.Emit("session_load", map[string]interface{}{"id": meta.ID, "title": meta.Title, "messages": len(loaded)})
		}
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
		if m.companionHub != nil {
			m.companionHub.Emit("auto_toggle", map[string]interface{}{"enabled": m.autoMode})
		}
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

	case "/version":
		m.msgs = append(m.msgs, ui.Message{
			Role: ui.RoleSystem, Content: fmt.Sprintf("택가이코드 (techai) %s", config.AppVersion), Timestamp: time.Now(),
		})
		m.updateViewport()
		return true, nil

	case "/multi":
		// ── SUSPENDED: sub-agent system globally disabled until re-open ──
		m.msgs = append(m.msgs, ui.Message{
			Role: ui.RoleSystem, Content: "[MULTI] ⏸ 멀티 에이전트 기능이 일시 중단되었습니다. 추후 재오픈 예정.", Timestamp: time.Now(),
		})
		m.updateViewport()
		return true, nil

	case "/companion":
		if m.companionHub == nil {
			m.companionHub = companion.NewHub()
		}
		if m.companionServer == nil {
			port := 8787
			if m.companionPort > 0 {
				port = m.companionPort
			}
			webFS, _ := fs.Sub(tgc.CompanionFS, "web")
			m.companionServer = companion.NewServer(m.companionHub, webFS, port)
			m.companionServer.Start()
			m.companionPort = port
			_ = companion.OpenBrowser(fmt.Sprintf("http://localhost:%d", port))
			m.msgs = append(m.msgs, ui.Message{
				Role: ui.RoleSystem, Content: fmt.Sprintf("[COMPANION] 브라우저 대시보드 시작 http://localhost:%d", port), Timestamp: time.Now(),
			})
			m.companionHub.Emit("state_snapshot", map[string]interface{}{
				"mode":      m.activeTab,
				"model":     m.currentModel(),
				"streaming": m.streaming,
				"autoMode":  m.autoMode,
				"sessionID": m.currentSessionID,
				"messages":  len(m.msgs),
			})
		} else {
			_ = companion.OpenBrowser(fmt.Sprintf("http://localhost:%d", m.companionPort))
			m.msgs = append(m.msgs, ui.Message{
				Role: ui.RoleSystem, Content: fmt.Sprintf("[COMPANION] 이미 실행 중 http://localhost:%d", m.companionPort), Timestamp: time.Now(),
			})
		}
		m.updateViewport()
		return true, nil


	case "/mcp":
		if m.mcpManager == nil {
			m.msgs = append(m.msgs, ui.Message{
				Role: ui.RoleSystem, Content: "[MCP] 설정된 서버 없음. config.yaml의 mcp.servers 확인", Timestamp: time.Now(),
			})
		} else {
			m.msgs = append(m.msgs, ui.Message{
				Role: ui.RoleSystem, Content: "[MCP] 서버 상태:\n" + m.mcpManager.Status(), Timestamp: time.Now(),
			})
		}
		m.updateViewport()
		return true, nil

	case "/copy":
		// /copy — last AI response, /copy N — Nth recent AI response, /copy all — entire session
		target := ""
		if arg == "all" {
			var sb strings.Builder
			for _, msg := range m.msgs {
				switch msg.Role {
				case ui.RoleUser:
					sb.WriteString("[사용자] " + msg.Content + "\n\n")
				case ui.RoleAssistant:
					sb.WriteString("[AI] " + msg.Content + "\n\n")
				}
			}
			target = sb.String()
		} else {
			// Find Nth AI response (default: last)
			n := 1
			if arg != "" {
				fmt.Sscanf(arg, "%d", &n)
			}
			count := 0
			for i := len(m.msgs) - 1; i >= 0; i-- {
				if m.msgs[i].Role == ui.RoleAssistant {
					count++
					if count == n {
						target = m.msgs[i].Content
						break
					}
				}
			}
		}
		if target == "" {
			m.msgs = append(m.msgs, ui.Message{Role: ui.RoleSystem, Content: "복사할 AI 응답이 없습니다.", Timestamp: time.Now()})
		} else {
			if err := clipboard.WriteAll(target); err != nil {
				m.msgs = append(m.msgs, ui.Message{Role: ui.RoleSystem, Content: fmt.Sprintf("클립보드 복사 실패: %v", err), Timestamp: time.Now()})
			} else {
				runes := []rune(target)
				preview := string(runes)
				if len(runes) > 60 {
					preview = string(runes[:60]) + "..."
				}
				m.msgs = append(m.msgs, ui.Message{Role: ui.RoleSystem, Content: fmt.Sprintf("클립보드에 복사됨 (%d자): %s", len(runes), preview), Timestamp: time.Now()})
			}
		}
		m.updateViewport()
		return true, nil

	case "/export":
		// Export session to markdown file
		filename := fmt.Sprintf("techai-session-%s.md", time.Now().Format("20060102-150405"))
		if arg != "" {
			filename = arg
			if !strings.HasSuffix(filename, ".md") {
				filename += ".md"
			}
		}
		var sb strings.Builder
		sb.WriteString(fmt.Sprintf("# TECHAI 세션 (%s)\n\n", time.Now().Format("2006-01-02 15:04")))
		for _, msg := range m.msgs {
			switch msg.Role {
			case ui.RoleUser:
				sb.WriteString(fmt.Sprintf("## 사용자\n\n%s\n\n", msg.Content))
			case ui.RoleAssistant:
				sb.WriteString(fmt.Sprintf("## AI\n\n%s\n\n", msg.Content))
			}
		}
		if err := os.WriteFile(filename, []byte(sb.String()), 0644); err != nil {
			m.msgs = append(m.msgs, ui.Message{Role: ui.RoleSystem, Content: fmt.Sprintf("내보내기 실패: %v", err), Timestamp: time.Now()})
		} else {
			m.msgs = append(m.msgs, ui.Message{Role: ui.RoleSystem, Content: fmt.Sprintf("세션 내보내기 완료: %s", filename), Timestamp: time.Now()})
		}
		m.updateViewport()
		return true, nil

	case "/diff":
		// Show git diff asynchronously to avoid blocking the TUI loop
		return true, func() tea.Msg {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			result, err := tools.ShellExec(ctx, "git diff")
			if err != nil {
				return slashResultMsg{content: fmt.Sprintf("diff 실패: %v", err)}
			}
			if result.Stdout == "" {
				return slashResultMsg{content: "변경사항 없음"}
			}
			diff := result.Stdout
			if len(diff) > 5000 {
				diff = diff[:5000] + "\n\n... (truncated)"
			}
			return slashResultMsg{content: diff}
		}

	case "/compact":
		// Manual context compaction
		beforeLen := len(m.history)
		beforeTokens := m.tokenCount
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()
		cap := llm.GetCapability(m.currentModel())
		target := cap.ContextWindow / 2 // compact to 50% of window
		m.history = llm.CompactWithLLM(ctx, m.client, m.currentModel(), m.history, target)
		afterLen := len(m.history)
		m.msgs = append(m.msgs, ui.Message{
			Role: ui.RoleSystem,
			Content: fmt.Sprintf("Compacted: %d → %d messages (tokens: %d → ~%d)",
				beforeLen, afterLen, beforeTokens, beforeTokens*afterLen/max(beforeLen, 1)),
			Timestamp: time.Now(),
		})
		m.updateViewport()
		return true, nil

	case "/init":
		mode := "simple"
		if arg == "deep" || arg == "hard" {
			mode = "deep"
		}

		// Step 1: Static analysis (both modes)
		profile := tools.GenerateProjectProfile(".")

		if mode == "deep" && m.client != nil {
			// Deep mode: LLM-powered analysis using Dev model
			m.msgs = append(m.msgs, ui.Message{Role: ui.RoleSystem, Content: "  /init deep — Analyzing project with AI...", Timestamp: time.Now()})
			m.updateViewport()

			// Collect key file contents for LLM analysis
			keyFiles := tools.CollectKeyFiles(".", 10)
			analysisPrompt := fmt.Sprintf(`Analyze this project and generate a comprehensive developer guide in Korean.

## Static Analysis Result:
%s

## Key File Contents:
%s

Write a comprehensive .techai.md that includes:
1. Project overview (한 줄 요약 + 상세 설명)
2. Architecture pattern (monorepo/monolith/microservice)
3. Key modules and their relationships
4. Important code patterns and conventions used
5. How to build, test, and deploy
6. Environment setup requirements
7. Common pitfalls and important notes

Format as clean markdown. Be specific, not generic. Reference actual file names and paths.`, profile, keyFiles)

			ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
			defer cancel()
			devModel := m.cfg.Models.Dev
			if devModel == "" {
				devModel = m.cfg.Models.Super
			}

			result, err := m.client.Chat(ctx, devModel, []openai.ChatCompletionMessage{
				{Role: openai.ChatMessageRoleUser, Content: analysisPrompt},
			})
			if err != nil {
				// LLM failed — fall back to static profile
				config.DebugLog("[INIT-DEEP] LLM analysis failed: %v", err)
				m.msgs = append(m.msgs, ui.Message{Role: ui.RoleSystem, Content: fmt.Sprintf("  Deep analysis failed (%v), using static profile.", err), Timestamp: time.Now()})
			} else {
				profile = result
			}
		}

		if err := os.WriteFile(".techai.md", []byte(profile), 0644); err != nil {
			m.msgs = append(m.msgs, ui.Message{Role: ui.RoleSystem, Content: fmt.Sprintf("Failed to write .techai.md: %v", err), Timestamp: time.Now()})
		} else {
			m.projectCtx = "\n\n## Project Context (.techai.md)\n" + profile
			if len(m.history) > 0 {
				md := llm.Mode(m.activeTab)
				m.history[0].Content = llm.SystemPrompt(md) + m.projectCtx
			}
			lines := strings.Count(profile, "\n")
			label := "simple"
			if mode == "deep" {
				label = "deep (AI-analyzed)"
			}
			m.msgs = append(m.msgs, ui.Message{
				Role:    ui.RoleSystem,
				Content: fmt.Sprintf(".techai.md generated [%s] (%d lines). Project context loaded.", label, lines),
				Timestamp: time.Now(),
			})
		}
		m.updateViewport()
		return true, nil

	case "/remember":
		if arg == "" {
			m.msgs = append(m.msgs, ui.Message{Role: ui.RoleSystem, Content: "Usage: /remember <text> | /remember -g <text> | /remember list | /remember edit <id> <text> | /remember delete <id> | /remember search <query>", Timestamp: time.Now()})
			m.updateViewport()
			return true, nil
		}
		switch {
		case arg == "list":
			m.msgs = append(m.msgs, ui.Message{Role: ui.RoleSystem, Content: m.memoryStore.List(), Timestamp: time.Now()})
		case strings.HasPrefix(arg, "search "):
			query := strings.TrimPrefix(arg, "search ")
			m.msgs = append(m.msgs, ui.Message{Role: ui.RoleSystem, Content: m.memoryStore.Search(query), Timestamp: time.Now()})
		case strings.HasPrefix(arg, "edit "):
			parts := strings.SplitN(strings.TrimPrefix(arg, "edit "), " ", 2)
			if len(parts) < 2 {
				m.msgs = append(m.msgs, ui.Message{Role: ui.RoleSystem, Content: "Usage: /remember edit <id> <new text>", Timestamp: time.Now()})
			} else {
				var id int
				fmt.Sscanf(parts[0], "%d", &id)
				result := m.memoryStore.Edit(id, parts[1], false)
				m.msgs = append(m.msgs, ui.Message{Role: ui.RoleSystem, Content: result, Timestamp: time.Now()})
			}
		case strings.HasPrefix(arg, "delete "):
			var id int
			fmt.Sscanf(strings.TrimPrefix(arg, "delete "), "%d", &id)
			result := m.memoryStore.Delete(id, false)
			m.msgs = append(m.msgs, ui.Message{Role: ui.RoleSystem, Content: result, Timestamp: time.Now()})
		case strings.HasPrefix(arg, "-g edit "):
			parts := strings.SplitN(strings.TrimPrefix(arg, "-g edit "), " ", 2)
			if len(parts) < 2 {
				m.msgs = append(m.msgs, ui.Message{Role: ui.RoleSystem, Content: "Usage: /remember -g edit <id> <new text>", Timestamp: time.Now()})
			} else {
				var id int
				fmt.Sscanf(parts[0], "%d", &id)
				result := m.memoryStore.Edit(id, parts[1], true)
				m.msgs = append(m.msgs, ui.Message{Role: ui.RoleSystem, Content: result, Timestamp: time.Now()})
			}
		case strings.HasPrefix(arg, "-g delete "):
			var id int
			fmt.Sscanf(strings.TrimPrefix(arg, "-g delete "), "%d", &id)
			result := m.memoryStore.Delete(id, true)
			m.msgs = append(m.msgs, ui.Message{Role: ui.RoleSystem, Content: result, Timestamp: time.Now()})
		case strings.HasPrefix(arg, "-g "):
			content := strings.TrimPrefix(arg, "-g ")
			result := m.memoryStore.Add(content, true)
			m.msgs = append(m.msgs, ui.Message{Role: ui.RoleSystem, Content: result, Timestamp: time.Now()})
		default:
			result := m.memoryStore.Add(arg, false)
			m.msgs = append(m.msgs, ui.Message{Role: ui.RoleSystem, Content: result, Timestamp: time.Now()})
		}
		m.updateViewport()
		return true, nil

	case "/forget":
		if arg == "" {
			m.msgs = append(m.msgs, ui.Message{Role: ui.RoleSystem, Content: "Usage: /forget <id> or /forget -g <id>", Timestamp: time.Now()})
		} else if strings.HasPrefix(arg, "-g ") {
			var id int
			fmt.Sscanf(strings.TrimPrefix(arg, "-g "), "%d", &id)
			m.msgs = append(m.msgs, ui.Message{Role: ui.RoleSystem, Content: m.memoryStore.Delete(id, true), Timestamp: time.Now()})
		} else {
			var id int
			fmt.Sscanf(arg, "%d", &id)
			m.msgs = append(m.msgs, ui.Message{Role: ui.RoleSystem, Content: m.memoryStore.Delete(id, false), Timestamp: time.Now()})
		}
		m.updateViewport()
		return true, nil

	case "/commands":
		if len(m.customCommands) == 0 {
			m.msgs = append(m.msgs, ui.Message{
				Role:      ui.RoleSystem,
				Content:   "No custom commands loaded. Add .md files to .tgc/commands/ or ~/.tgc/commands/",
				Timestamp: time.Now(),
			})
		} else {
			var lines []string
			lines = append(lines, "  Custom commands:")
			for name := range m.customCommands {
				lines = append(lines, fmt.Sprintf("  /%s", name))
			}
			m.msgs = append(m.msgs, ui.Message{
				Role:      ui.RoleSystem,
				Content:   strings.Join(lines, "\n"),
				Timestamp: time.Now(),
			})
		}
		m.updateViewport()
		return true, nil

	case "/exit", "/quit":
		return true, tea.Quit

	case "/undo":
		count := 1
		if arg != "" {
			if arg == "list" {
				list := tools.FormatSnapshotList(20)
				m.msgs = append(m.msgs, ui.Message{
					Role: ui.RoleSystem, Content: "  스냅샷 목록:\n" + list, Timestamp: time.Now(),
				})
				m.updateViewport()
				return true, nil
			}
			fmt.Sscanf(arg, "%d", &count)
		}
		result := tools.UndoLast(count)
		m.msgs = append(m.msgs, ui.Message{
			Role: ui.RoleSystem, Content: result, Timestamp: time.Now(),
		})
		m.updateViewport()
		return true, nil

	case "/help":
		help := fmt.Sprintf(`  TECHAI CODE %s
  Enter — Send    Ctrl+J — Newline
  Ctrl+K — Palette    Esc — Menu    Ctrl+Y — Copy last reply    Ctrl+C — Quit
  ↑/↓ — Scroll (input empty) / History    Alt+↑/↓ — Scroll    PgUp/PgDn — Page scroll
  Drag — Select & copy text    /copy — Copy last AI reply    /copy all — Copy session

  /init — Quick scan    /init deep — AI-powered deep analysis
  /remember <text> — Save memory    /remember list — Show all
  /remember -g <text> — Global memory    /forget <id> — Delete
  /commands — List custom commands (.tgc/commands/)
  /new — New session    /sessions — List    /session <id> — Restore
  /auto — Auto mode    /multi — Multi-agent    /diagnostics — Lint
  /git — Git status    /diff — Git changes    /version — Version
  /copy — Copy AI response    /export — Export session to .md
  /undo — Undo file edit    /undo list — Snapshot history
  /companion — Browser dashboard    /mcp — MCP server status
  /compact — Compress history    /clear — Clear chat
  /commands — List custom commands    /setup — Reset config    /exit — Quit`, config.AppVersion)
		m.msgs = append(m.msgs, ui.Message{
			Role: ui.RoleSystem, Content: help, Timestamp: time.Now(),
		})
		m.updateViewport()
		return true, nil
	}

	// Check custom commands
	cmdName := strings.TrimPrefix(cmd, "/")
	if template, ok := m.customCommands[cmdName]; ok {
		message := strings.ReplaceAll(template, "$ARGUMENTS", arg)
		config.DebugLog("[COMMANDS] executing custom command /%s with arg=%q", cmdName, arg)
		return true, m.sendMessage(message)
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
		}
		// Apply text selection highlighting
		if m.selecting || (m.selStartX != m.selEndX || m.selStartY != m.selEndY) {
			contentLines = m.highlightSelection(contentLines)
		}
		vpContent = strings.Join(contentLines, "\n")

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
		statusBar := ui.RenderStatusBar(displayModel, m.tokenCount, ctxWindow, elapsed, m.activeTab, m.cwd, m.width, config.IsDebug(), len(tools.ToolsForMode(m.activeTab)), m.gitInfo.Label(), m.multiEnabled)

		// Overlay palette or menu on top of the viewport when active.
		if m.showSessionPicker {
			overlay := ui.RenderSessionPicker(m.sessionPickerItems, m.sessionPickerSelected, m.currentSessionID, m.width)
			vpContent = lipgloss.Place(m.width, m.viewport.Height(), lipgloss.Center, lipgloss.Center, overlay)
		} else if m.showPalette {
			overlay := ui.RenderPalette(m.paletteFiltered, m.paletteSelected, m.paletteQuery, m.width)
			vpContent = lipgloss.Place(m.width, m.viewport.Height(), lipgloss.Center, lipgloss.Center, overlay)
		} else if m.showMenu {
			overlay := ui.RenderMenu(ui.MainMenuItems, m.menuSelected, m.width)
			vpContent = lipgloss.Place(m.width, m.viewport.Height(), lipgloss.Center, lipgloss.Center, overlay)
		}

		// Show paste hint above input box if present
		if m.pasteHint != "" {
			hintStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FBBF24")).Italic(true)
			content = lipgloss.JoinVertical(lipgloss.Left, vpContent, hintStyle.Render("  "+m.pasteHint), inputBox, statusBar)
		} else {
			content = lipgloss.JoinVertical(lipgloss.Left, vpContent, inputBox, statusBar)
		}
	}

	v := tea.NewView(content)
	v.AltScreen = true
	v.MouseMode = tea.MouseModeCellMotion
	v.KeyboardEnhancements.ReportEventTypes = true
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

// --- Text selection helpers ---

func (m *Model) clearSelection() {
	m.selecting = false
	m.selStartX, m.selStartY = 0, 0
	m.selEndX, m.selEndY = 0, 0
}

func (m *Model) extractSelectedText() string {
	content := m.viewport.GetContent()
	plain := ansi.Strip(content)
	lines := strings.Split(plain, "\n")

	startY, endY := m.selStartY, m.selEndY
	startX, endX := m.selStartX, m.selEndX

	// Normalize: ensure start is before end
	if startY > endY || (startY == endY && startX > endX) {
		startY, endY = endY, startY
		startX, endX = endX, startX
	}

	if startY >= len(lines) {
		return ""
	}
	if endY >= len(lines) {
		endY = len(lines) - 1
		endX = len([]rune(lines[endY]))
	}

	if startY == endY {
		// Single line selection
		runes := []rune(lines[startY])
		sx := min(startX, len(runes))
		ex := min(endX, len(runes))
		if sx == ex {
			return ""
		}
		return string(runes[sx:ex])
	}

	// Multi-line selection
	var sb strings.Builder
	// First line: from startX to end
	runes := []rune(lines[startY])
	sx := min(startX, len(runes))
	sb.WriteString(string(runes[sx:]))
	sb.WriteByte('\n')

	// Middle lines: full content
	for y := startY + 1; y < endY; y++ {
		if y < len(lines) {
			sb.WriteString(lines[y])
			sb.WriteByte('\n')
		}
	}

	// Last line: from start to endX
	runes = []rune(lines[endY])
	ex := min(endX, len(runes))
	sb.WriteString(string(runes[:ex]))
	return sb.String()
}

// highlightSelection applies reverse-video styling to the selected region
// of the visible viewport lines. Coordinates are in content-space (YOffset-based).
func (m *Model) highlightSelection(visibleLines []string) []string {
	startY, endY := m.selStartY, m.selEndY
	startX, endX := m.selStartX, m.selEndX
	if startY > endY || (startY == endY && startX > endX) {
		startY, endY = endY, startY
		startX, endX = endX, startX
	}

	yOff := m.viewport.YOffset()
	result := make([]string, len(visibleLines))
	copy(result, visibleLines)

	for i, line := range result {
		contentY := yOff + i
		if contentY < startY || contentY > endY {
			continue
		}
		lineWidth := ansi.StringWidth(line)

		var sx, ex int
		if contentY == startY {
			sx = startX
		}
		if contentY == endY {
			ex = endX
		} else {
			ex = lineWidth
		}
		if sx >= lineWidth {
			continue
		}
		if ex > lineWidth {
			ex = lineWidth
		}
		if sx >= ex {
			continue
		}

		before := ansi.Cut(line, 0, sx)
		sel := ansi.Cut(line, sx, ex)
		after := ansi.Cut(line, ex, lineWidth)

		// Apply reverse video to selected portion
		highlighted := lipgloss.NewStyle().Reverse(true).Render(ansi.Strip(sel))
		result[i] = before + highlighted + after
	}
	return result
}

var spinnerFrames = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}

func (m *Model) streamStatus() string {
	elapsed := time.Since(m.streamStart)
	frame := spinnerFrames[int(elapsed.Milliseconds()/150)%len(spinnerFrames)]

	// Multi-agent progress display
	if m.multiRunning && len(m.multiProgress) > 0 {
		var lines []string
		for _, p := range m.multiProgress {
			icon := "●"
			if p.Status == "done" {
				icon = "✓"
			} else if p.Status == "waiting" {
				icon = "○"
			} else if p.Status == "error" {
				icon = "✗"
			} else if p.Status == "synthesizing" {
				icon = "⚡"
			}
			detail := p.Status
			if p.Detail != "" {
				detail = p.Detail
			}
			line := fmt.Sprintf("  %s %s  %s  %dtok  (%.1fs)",
				icon, p.Agent, detail, p.Tokens, p.Elapsed.Seconds())
			lines = append(lines, line)
		}
		return fmt.Sprintf("%s Multi 실행중 (%.1fs)\n%s", frame, elapsed.Seconds(), strings.Join(lines, "\n"))
	}

	if m.multiRunning {
		return fmt.Sprintf("%s Multi 시작중... (%.1fs)", frame, elapsed.Seconds())
	}

	if m.lastChunkAt.IsZero() {
		// Show warning in chat area after 15s of no response (once)
		if elapsed > 15*time.Second && !m.streamWarnShown {
			m.streamWarnShown = true
			m.msgs = append(m.msgs, ui.Message{
				Role: ui.RoleSystem, Content: "  Waiting for response... (server may be slow, press Esc to cancel)", Timestamp: time.Now(), Tag: "stream-warn",
			})
		}
		if elapsed > 30*time.Second {
			return fmt.Sprintf("%s Connecting... (%.0fs — no response)", frame, elapsed.Seconds())
		}
		return fmt.Sprintf("%s Connecting... (%.1fs)", frame, elapsed.Seconds())
	}

	sinceLastChunk := time.Since(m.lastChunkAt)

	// Show stall warning in chat area after 15s of silence mid-stream (once)
	if sinceLastChunk > 15*time.Second && !m.streamWarnShown {
		m.streamWarnShown = true
		m.msgs = append(m.msgs, ui.Message{
			Role: ui.RoleSystem, Content: "  Response stalled... (press Esc to cancel and retry)", Timestamp: time.Now(), Tag: "stream-warn",
		})
	}
	tps := float64(0)
	if elapsed.Seconds() > 0 {
		tps = float64(m.tokenCount) / elapsed.Seconds()
	}

	if sinceLastChunk > 15*time.Second {
		return fmt.Sprintf("%s Stalled (%.0fs waiting · %dtok)", frame, sinceLastChunk.Seconds(), m.tokenCount)
	}
	if sinceLastChunk > 2*time.Second {
		return fmt.Sprintf("%s Thinking... (%.0fs · %dtok · %.1ftok/s)", frame, elapsed.Seconds(), m.tokenCount, tps)
	}
	if m.wasThinking {
		return fmt.Sprintf("%s 💭 Reasoning (%.1fs · %dtok · %.1ftok/s)", frame, elapsed.Seconds(), m.tokenCount, tps)
	}
	return fmt.Sprintf("%s ✍️ Writing (%.1fs · %dtok · %.1ftok/s)", frame, elapsed.Seconds(), m.tokenCount, tps)
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

	// Inject auto-prefetched file contents into the last user message
	// in the COPY only — original history stays clean to prevent bloat.
	// Keep pendingPrefetch alive across tool iterations so the model always
	// sees the file contents. It gets replaced on the next sendMessage call.
	if m.pendingPrefetch != "" && len(history) > 0 {
		for i := len(history) - 1; i >= 0; i-- {
			if history[i].Role == openai.ChatMessageRoleUser {
				history[i].Content += m.pendingPrefetch
				break
			}
		}
	}

	toolDefs := tools.ToolsForMode(m.activeTab)

	// When auto-prefetch is active, remove apply_patch and file_edit from tools.
	// The model already has full file contents, so file_write (full rewrite) is
	// faster and more reliable than trying to generate correct patch hunks.
	if m.pendingPrefetch != "" {
		filtered := make([]openai.Tool, 0, len(toolDefs))
		for _, td := range toolDefs {
			if td.Function != nil && (td.Function.Name == "apply_patch" || td.Function.Name == "file_edit" || td.Function.Name == "hashline_edit") {
				continue
			}
			filtered = append(filtered, td)
		}
		toolDefs = filtered
		config.DebugLog("[PREFETCH] removed apply_patch/file_edit from tools — forcing file_write")
	}

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
	// Reset search state + extract keywords from user message
	tools.ResetFailedPatterns()
	tools.SetUserContext(input)

	m.msgs = append(m.msgs, ui.Message{
		Role: ui.RoleUser, Content: input, Timestamp: time.Now(),
	})
	if m.companionHub != nil {
		m.companionHub.Emit("user_message", map[string]interface{}{"content": input})
	}

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

	// Auto-prefetch: detect file paths in user input and inject file contents.
	// Contents are stored as a separate "prefetch" message that is excluded
	// from persistent history — this prevents context bloat across turns.
	prefetched := autoPrefetchFiles(input)
	if prefetched != "" {
		config.DebugLog("[PREFETCH] injected %d chars of file contents", len(prefetched))
		m.pendingPrefetch = prefetched
	} else {
		m.pendingPrefetch = ""
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
	m.streamRetries = 0
	m.wasThinking = false
	m.streamStart = time.Now()
	m.lastChunkAt = time.Time{}
	m.streamWarnShown = false
	m.updateViewport()
	if m.companionHub != nil {
		m.companionHub.Emit("stream_start", map[string]interface{}{"model": m.currentModel(), "mode": m.activeTab})
	}

	// Check if multi-agent should handle this request
	if m.shouldUseMulti(input) {
		return m.startMultiStream(input)
	}

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
	m.streamStart = time.Now()
	m.lastChunkAt = time.Time{}
	m.streamWarnShown = false
	return m.startStream()
}

func (m *Model) waitForNextChunk() tea.Cmd {
	ch := m.streamCh
	if ch == nil {
		return nil
	}
	return func() tea.Msg {
		// Timeout: if no chunk arrives in 30 seconds, return error
		// so the app can show warning / retry
		select {
		case chunk, ok := <-ch:
			if !ok {
				return streamChunkMsg{done: true}
			}
			return streamChunkMsg{
				content:    chunk.Content,
				isThinking: chunk.IsThinking,
				done:       chunk.Done,
				err:        chunk.Err,
				toolCalls:  chunk.ToolCalls,
				usage:      chunk.Usage,
			}
		case <-time.After(45 * time.Second):
			return streamChunkMsg{
				err: fmt.Errorf("no response from server (45s timeout). Press Enter to retry"),
			}
		}
	}
}

func (m *Model) cancelMulti() {
	if m.multiCancel != nil {
		m.multiCancel()
		m.multiCancel = nil
	}
	m.multiRunning = false
	m.multiProgress = nil
	m.multiProgressCh = nil
	m.multiResultCh = nil
	m.streaming = false
	m.lastElapsed = time.Since(m.streamStart)
	m.msgs = append(m.msgs, ui.Message{
		Role: ui.RoleSystem, Content: "[Multi] 중단됨", Timestamp: time.Now(),
	})
	m.updateViewport()
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

// waitForNextMulti returns a tea.Cmd that polls for the next multi event
// from the stored progress/result channels on the Model.
func (m *Model) waitForNextMulti() tea.Cmd {
	if !m.multiRunning {
		return nil
	}
	progressCh := m.multiProgressCh
	resultCh := m.multiResultCh
	return func() tea.Msg {
		select {
		case p, ok := <-progressCh:
			if !ok {
				// Progress closed, wait for final result
				result := <-resultCh
				return multiResultMsg{result: result}
			}
			return multiProgressMsg{progress: p}
		case result := <-resultCh:
			return multiResultMsg{result: result}
		}
	}
}

// shouldUseMulti decides whether to activate multi-agent for this input.
// Hybrid approach: keyword matching (instant) → LLM fallback (5s timeout).
func (m *Model) shouldUseMulti(_ string) bool {
	// ── SUSPENDED: sub-agent system globally disabled until re-open ──
	// All multi-agent functionality is preserved in code but hard-gated here.
	// To re-enable: restore the original logic from git history.
	return false
}

// startMultiStream launches the multi-agent orchestrator in a background goroutine.
func (m *Model) startMultiStream(input string) tea.Cmd {
	ctx, cancel := context.WithCancel(context.Background())
	m.multiCancel = cancel
	m.multiRunning = true
	m.multiProgress = nil
	m.streamStart = time.Now()

	// Determine strategy
	strategy := m.multiStrategy
	if m.multiAuto {
		strategy = multi.AutoDetectStrategy(input)
	}

	config.DebugLog("[APP-MULTI] start strategy=%s model1=%s model2=%s",
		strategy, m.cfg.Models.Super, m.cfg.Models.Dev)

	// Show activation message
	m.msgs = append(m.msgs, ui.Message{
		Role: ui.RoleSystem,
		Content: fmt.Sprintf("  [Multi:%s] %s + %s",
			strategy, shortModel(m.cfg.Models.Super), shortModel(m.cfg.Models.Dev)),
		Timestamp: time.Now(),
	})
	m.updateViewport()

	workDir, _ := os.Getwd()
	orch := multi.NewOrchestrator(m.client, m.cfg.Models.Super, m.cfg.Models.Dev, workDir)

	// Store channels on Model so waitForNextMulti can access them
	m.multiProgressCh = orch.Progress()

	// Copy history for the orchestrator
	history := make([]openai.ChatCompletionMessage, len(m.history))
	copy(history, m.history)

	mode := m.activeTab

	// Launch orchestrator in background, pipe result through a channel
	resultCh := make(chan multi.MergedResult, 1)
	m.multiResultCh = resultCh
	go func() {
		resultCh <- orch.Run(ctx, strategy, history, mode)
	}()

	// Return the first wait command
	return m.waitForNextMulti()
}

func shortModel(model string) string {
	if idx := strings.LastIndex(model, "/"); idx >= 0 {
		return model[idx+1:]
	}
	return model
}

func truncate(s string, n int) string {
	runes := []rune(s)
	if len(runes) <= n {
		return s
	}
	return string(runes[:n]) + "..."
}

// formatToolCallPreview returns a user-friendly tool call description.
func formatToolCallPreview(name, argsJSON string) string {
	// Extract key argument for display
	var args map[string]interface{}
	_ = json.Unmarshal([]byte(argsJSON), &args)

	icons := map[string]string{
		"file_read":        "reading",
		"file_write":       "writing",
		"file_edit":        "editing",
		"apply_patch":      "patching",
		"list_files":       "listing",
		"shell_exec":       "running",
		"grep_search":      "searching",
		"glob_search":      "finding",
		"hashline_read":    "reading",
		"hashline_edit":    "editing",
		"git_status":       "git status",
		"git_diff":         "git diff",
		"git_log":          "git log",
		"diagnostics":      "diagnosing",
		"knowledge_search": "searching docs",
	}

	action := icons[name]
	if action == "" {
		action = name
	}

	// Show the most relevant argument
	detail := ""
	switch name {
	case "file_read", "file_write", "file_edit", "hashline_read", "hashline_edit":
		if p, ok := args["path"].(string); ok {
			detail = p
		}
	case "shell_exec":
		if c, ok := args["command"].(string); ok {
			if len(c) > 60 {
				c = c[:60] + "..."
			}
			detail = c
		}
	case "grep_search":
		if p, ok := args["pattern"].(string); ok {
			detail = p
		}
	case "glob_search":
		if p, ok := args["pattern"].(string); ok {
			detail = p
		}
	case "knowledge_search":
		if q, ok := args["query"].(string); ok {
			detail = q
		}
	default:
		detail = truncateArgs(argsJSON, 50)
	}

	if detail != "" {
		return fmt.Sprintf(">> %s %s", action, detail)
	}
	return fmt.Sprintf(">> %s", action)
}

// formatToolResultPreview returns a rich preview of tool execution results.
// Shows code snippets for file reads, diff for edits, match lines for grep.
func formatToolResultPreview(name, output string) string {
	if strings.HasPrefix(output, "Error:") {
		return fmt.Sprintf("<< %s: %s", name, truncateArgs(output, 120))
	}

	switch name {
	case "file_read", "hashline_read":
		return formatFileReadPreview(output)

	case "grep_search":
		return formatGrepPreview(output)

	case "glob_search":
		return formatGlobPreview(output)

	case "file_edit", "hashline_edit":
		return formatEditPreview(output)

	case "apply_patch":
		return formatPatchPreview(output)

	case "list_files":
		return formatListPreview(output)

	case "shell_exec":
		lines := strings.Split(strings.TrimSpace(output), "\n")
		if len(lines) <= 4 {
			return "<< " + name + ":\n" + indentPreview(output, 3)
		}
		preview := strings.Join(lines[:3], "\n") + fmt.Sprintf("\n... (%d lines total)", len(lines))
		return "<< " + name + ":\n" + indentPreview(preview, 3)

	default:
		return fmt.Sprintf("<< %s: %s", name, truncateArgs(output, 120))
	}
}

func formatFileReadPreview(output string) string {
	lines := strings.Split(output, "\n")
	total := len(lines)

	// Show first 5 lines + total count
	showLines := 5
	if total < showLines {
		showLines = total
	}
	preview := strings.Join(lines[:showLines], "\n")
	if total > showLines {
		preview += fmt.Sprintf("\n... (%d lines)", total)
	}
	return "<< file_read:\n" + indentPreview(preview, 3)
}

func formatGrepPreview(output string) string {
	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) == 0 || (len(lines) == 1 && lines[0] == "") {
		return "<< grep_search: No matches found."
	}

	// Show up to 6 match lines
	showLines := 6
	if len(lines) < showLines {
		showLines = len(lines)
	}
	preview := strings.Join(lines[:showLines], "\n")
	if len(lines) > showLines {
		preview += fmt.Sprintf("\n... (%d matches total)", len(lines))
	}
	return "<< grep_search:\n" + indentPreview(preview, 3)
}

func formatGlobPreview(output string) string {
	files := strings.Split(strings.TrimSpace(output), "\n")
	if len(files) == 0 || (len(files) == 1 && files[0] == "") {
		return "<< glob_search: No files matched."
	}
	if len(files) <= 5 {
		return "<< glob_search: " + strings.Join(files, " ")
	}
	return fmt.Sprintf("<< glob_search: %s ... (%d files)", strings.Join(files[:4], " "), len(files))
}

func formatEditPreview(output string) string {
	if strings.Contains(output, "applied successfully") || strings.Contains(output, "OK") {
		// Try to show what changed
		return "<< file_edit: " + truncateArgs(output, 150)
	}
	return "<< file_edit: " + truncateArgs(output, 150)
}

func formatPatchPreview(output string) string {
	lines := strings.Split(strings.TrimSpace(output), "\n")
	// Show summary line + first few diff lines (bounded)
	var summary []string
	var diffLines []string
	for _, line := range lines {
		if (strings.HasPrefix(line, "~") || strings.HasPrefix(line, "Errors:")) && len(summary) < 5 {
			summary = append(summary, line)
		} else if strings.HasPrefix(line, "---") || strings.HasPrefix(line, "+++") || strings.HasPrefix(line, "@@") {
			if len(diffLines) < 12 {
				diffLines = append(diffLines, line)
			}
		} else if (strings.HasPrefix(line, "-") || strings.HasPrefix(line, "+")) && len(diffLines) < 12 {
			diffLines = append(diffLines, line)
		}
	}
	result := "<< apply_patch: " + strings.Join(summary, " ")
	if len(diffLines) > 0 {
		showLines := 8
		if len(diffLines) < showLines {
			showLines = len(diffLines)
		}
		result += "\n" + indentPreview(strings.Join(diffLines[:showLines], "\n"), 3)
	}
	return result
}

func formatListPreview(output string) string {
	items := strings.Split(strings.TrimSpace(output), "\n")
	if len(items) <= 8 {
		return "<< list_files: " + strings.Join(items, " ")
	}
	return fmt.Sprintf("<< list_files: %s ... (%d items)", strings.Join(items[:6], " "), len(items))
}

func indentPreview(s string, spaces int) string {
	prefix := strings.Repeat(" ", spaces)
	lines := strings.Split(s, "\n")
	for i, line := range lines {
		lines[i] = prefix + line
	}
	return strings.Join(lines, "\n")
}

func truncateArgs(s string, max int) string {
	s = strings.ReplaceAll(s, "\n", " ")
	runes := []rune(s)
	if len(runes) > max {
		return string(runes[:max]) + "..."
	}
	return s
}

// stripTrailingNonPath truncates anything after a recognized file extension.
// Korean particles, punctuation, or any non-path characters are stripped.
// "src/views/HomePage.tsx의" → "src/views/HomePage.tsx"
// "src/data/mock.ts를"      → "src/data/mock.ts"
// "README.md에서"           → "README.md"
func stripTrailingNonPath(s string) string {
	exts := []string{
		".tsx", ".ts", ".jsx", ".js", ".mjs", ".cjs",
		".go", ".py", ".java", ".rs", ".rb", ".php",
		".css", ".scss", ".html", ".vue", ".svelte",
		".json", ".yaml", ".yml", ".toml", ".xml",
		".md", ".txt", ".sql", ".sh", ".env",
		".prisma", ".graphql", ".proto",
	}
	for _, ext := range exts {
		idx := strings.Index(s, ext)
		if idx >= 0 {
			return s[:idx+len(ext)]
		}
	}
	return s
}

// autoPrefetchFiles detects file paths in user input (e.g. "src/views/HomePage.tsx")
// and auto-reads them, injecting the contents into the prompt so the model can
// skip file_read tool calls and go straight to editing.
func autoPrefetchFiles(input string) string {
	// Match patterns like src/..., ./..., or any path with / and a file extension
	words := strings.Fields(input)
	var paths []string
	seen := make(map[string]bool)

	for _, word := range words {
		// Clean punctuation from word edges
		w := strings.Trim(word, ".,;:!?\"'`()[]{}")
		// Truncate after file extension — strips Korean particles or any trailing text.
		// "src/views/HomePage.tsx의" → "src/views/HomePage.tsx"
		w = stripTrailingNonPath(w)
		// Must contain / and look like a file path
		if !strings.Contains(w, "/") {
			continue
		}
		// Must have a file extension
		if !strings.Contains(filepath.Base(w), ".") {
			continue
		}
		// Skip URLs
		if strings.HasPrefix(w, "http") {
			continue
		}
		// Normalize: remove leading ./
		w = strings.TrimPrefix(w, "./")
		if !seen[w] {
			seen[w] = true
			paths = append(paths, w)
		}
	}

	if len(paths) == 0 {
		return ""
	}

	// Limit to 4 files max to prevent context overflow
	if len(paths) > 4 {
		paths = paths[:4]
	}

	var sb strings.Builder
	sb.WriteString("\n\n---\n[자동 첨부된 파일 내용 — 도구 호출 없이 바로 수정하세요]\n\n")

	totalBytes := 0
	const maxTotalBytes = 50000 // 50KB cap to prevent context overflow

	for _, p := range paths {
		data, err := os.ReadFile(p)
		if err != nil {
			// Fallback: search by filename in common directories
			basename := filepath.Base(p)
			found := false
			for _, dir := range []string{"src", ".", "app", "lib", "components", "pages", "views"} {
				matches, _ := filepath.Glob(filepath.Join(dir, "**", basename))
				if len(matches) == 0 {
					// Try one-level deeper
					matches, _ = filepath.Glob(filepath.Join(dir, "*", basename))
				}
				if len(matches) == 0 {
					matches, _ = filepath.Glob(filepath.Join(dir, "*", "*", basename))
				}
				if len(matches) > 0 {
					data, err = os.ReadFile(matches[0])
					if err == nil {
						p = matches[0]
						found = true
						config.DebugLog("[PREFETCH] fallback found %s at %s", basename, p)
						break
					}
				}
			}
			if !found {
				continue
			}
		}
		content := string(data)
		lines := strings.Split(content, "\n")
		// Skip files > 300 lines
		if len(lines) > 300 {
			sb.WriteString(fmt.Sprintf("### %s (%d lines — 너무 길어 생략, file_read로 읽으세요)\n\n", p, len(lines)))
			continue
		}
		// Cumulative byte cap
		if totalBytes+len(content) > maxTotalBytes {
			sb.WriteString(fmt.Sprintf("### %s (용량 초과 — file_read로 읽으세요)\n\n", p))
			continue
		}
		totalBytes += len(content)
		sb.WriteString(fmt.Sprintf("### %s (%d lines)\n```\n%s\n```\n\n", p, len(lines), content))
	}

	// Add instruction to use file_write for the edit
	sb.WriteString("[중요: 위 파일을 수정할 때 file_write로 전체 파일을 한 번에 교체하세요. apply_patch보다 안전합니다.]\n")

	return sb.String()
}

// parseKnowledgePacks extracts knowledge_packs from .techai.md content.
// Looks for a line like: "- **Knowledge packs**: react, database, css"
// or "knowledge_packs: react, database, css"
func parseKnowledgePacks(projectCtx string) []string {
	for _, line := range strings.Split(projectCtx, "\n") {
		lower := strings.ToLower(strings.TrimSpace(line))

		// Match "knowledge_packs: ..." or "- **Knowledge packs**: ..."
		var value string
		if strings.HasPrefix(lower, "knowledge_packs:") {
			value = strings.TrimSpace(line[len("knowledge_packs:"):])
		} else if strings.Contains(lower, "knowledge packs") && strings.Contains(line, ":") {
			idx := strings.LastIndex(line, ":")
			value = strings.TrimSpace(line[idx+1:])
		}

		if value != "" {
			// Strip markdown bold markers
			value = strings.ReplaceAll(value, "**", "")
			var packs []string
			for _, p := range strings.Split(value, ",") {
				p = strings.TrimSpace(p)
				if p != "" && p != "none" {
					packs = append(packs, p)
				}
			}
			return packs
		}
	}
	return nil
}

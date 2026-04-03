package app

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	openai "github.com/sashabaranov/go-openai"

	"github.com/kimjiwon/tgc/internal/config"
	"github.com/kimjiwon/tgc/internal/llm"
	"github.com/kimjiwon/tgc/internal/ui"
)

type streamChunkMsg struct {
	content string
	done    bool
	err     error
}


type Model struct {
	cfg       config.Config
	client    *llm.Client
	activeTab int

	histories [llm.ModeCount][]openai.ChatCompletionMessage
	messages  [llm.ModeCount][]ui.Message

	textarea textarea.Model
	viewport viewport.Model

	streaming    bool
	streamBuf    string
	streamCh     <-chan llm.StreamChunk
	streamCancel context.CancelFunc
	streamStart  time.Time
	lastElapsed  time.Duration
	tokenCount   int

	inSetup    bool
	setupInput textarea.Model
	setupCfg   config.Config

	width  int
	height int
	ready  bool
}

func NewModel(cfg config.Config, initialMode int, needsSetup bool) Model {
	ta := textarea.New()
	ta.Placeholder = "Type your message..."
	ta.CharLimit = 4096
	ta.SetWidth(80)
	ta.SetHeight(1)
	ta.Focus()
	ta.ShowLineNumbers = false

	setupTa := textarea.New()
	setupTa.Placeholder = "sk_..."
	setupTa.CharLimit = 512
	setupTa.SetWidth(60)
	setupTa.SetHeight(1)
	setupTa.ShowLineNumbers = false

	vp := viewport.New(80, 20)

	m := Model{
		cfg:        cfg,
		activeTab:  initialMode,
		textarea:   ta,
		viewport:   vp,
		inSetup:    needsSetup,
		setupCfg:   config.DefaultConfig(),
		setupInput: setupTa,
	}

	if needsSetup {
		m.setupInput.Focus()
	} else {
		m.client = llm.NewClient(cfg.API.BaseURL, cfg.API.APIKey)
	}

	for i := 0; i < llm.ModeCount; i++ {
		mode := llm.Mode(i)
		m.histories[i] = []openai.ChatCompletionMessage{
			{Role: openai.ChatMessageRoleSystem, Content: llm.SystemPrompt(mode)},
		}
		m.messages[i] = []ui.Message{
			{Role: ui.RoleSystem, Content: ui.ModeWelcome(i), Timestamp: time.Now()},
		}
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
	case tea.KeyMsg:
		if m.streaming {
			if msg.Type == tea.KeyCtrlC || msg.Type == tea.KeyEsc {
				m.cancelStream()
			}
			return m, nil
		}

		switch msg.Type {
		case tea.KeyCtrlC:
			return m, tea.Quit

		case tea.KeyCtrlL:
			m.messages[m.activeTab] = m.messages[m.activeTab][:1]
			m.histories[m.activeTab] = m.histories[m.activeTab][:1]
			m.streamBuf = ""
			m.tokenCount = 0
			m.lastElapsed = 0
			m.updateViewport()
			return m, nil

		case tea.KeyTab:
			m.activeTab = (m.activeTab + 1) % llm.ModeCount
			m.updateViewport()
			return m, nil

		case tea.KeyEnter:
			if msg.Alt {
				// Shift/Alt+Enter = newline, grow textarea
				h := m.textarea.Height()
				if h < 6 {
					m.textarea.SetHeight(h + 1)
					m.recalcLayout()
				}
				break
			}
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

		case tea.KeyPgUp, tea.KeyPgDown, tea.KeyUp, tea.KeyDown:
			// Pass scroll keys to viewport when textarea is single line
			if m.textarea.Height() <= 1 && (msg.Type == tea.KeyPgUp || msg.Type == tea.KeyPgDown) {
				var vpCmd tea.Cmd
				m.viewport, vpCmd = m.viewport.Update(msg)
				return m, vpCmd
			}
		}

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
			m.messages[m.activeTab] = append(m.messages[m.activeTab], ui.Message{
				Role: ui.RoleSystem, Content: fmt.Sprintf("오류: %v", msg.err), Timestamp: time.Now(),
			})
			m.streamBuf = ""
			m.updateViewport()
			return m, nil
		}
		if msg.done {
			m.streaming = false
			m.streamCh = nil
			m.lastElapsed = time.Since(m.streamStart)
			if m.streamBuf != "" {
				m.messages[m.activeTab] = append(m.messages[m.activeTab], ui.Message{
					Role: ui.RoleAssistant, Content: m.streamBuf, Timestamp: time.Now(),
				})
				m.histories[m.activeTab] = append(m.histories[m.activeTab], openai.ChatCompletionMessage{
					Role: openai.ChatMessageRoleAssistant, Content: m.streamBuf,
				})
			}
			m.streamBuf = ""
			m.updateViewport()
			return m, nil
		}
		m.streamBuf += msg.content
		m.tokenCount++
		m.updateViewport()
		return m, m.waitForNextChunk()
	}

	// Only pass key events to textarea, not mouse events
	if _, ok := msg.(tea.KeyMsg); ok {
		var taCmd tea.Cmd
		m.textarea, taCmd = m.textarea.Update(msg)
		return m, taCmd
	}

	return m, nil
}

func (m Model) updateSetup(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC:
			return m, tea.Quit
		case tea.KeyEnter:
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
		m.setupInput.Placeholder = "sk_..."
		m.setupInput.Focus()
		return true, nil

	case "/clear":
		m.messages[m.activeTab] = m.messages[m.activeTab][:1]
		m.histories[m.activeTab] = m.histories[m.activeTab][:1]
		m.streamBuf = ""
		m.tokenCount = 0
		m.lastElapsed = 0
		m.updateViewport()
		return true, nil

	case "/help":
		help := `  /clear — 대화삭제    Ctrl+C — 종료`
		m.messages[m.activeTab] = append(m.messages[m.activeTab], ui.Message{
			Role: ui.RoleSystem, Content: help, Timestamp: time.Now(),
		})
		m.updateViewport()
		return true, nil
	}
	return false, nil
}

func (m Model) View() string {
	if m.inSetup {
		return m.viewSetup()
	}
	if !m.ready {
		return "\n  로딩중..."
	}

	content := m.viewport.View()

	inputBox := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#334155")).
		Width(m.width - 4).
		Render(m.textarea.View())

	model := m.currentModel()
	elapsed := m.lastElapsed
	if m.streaming {
		elapsed = time.Since(m.streamStart)
	}
	statusBar := ui.RenderStatusBar(model, m.tokenCount, elapsed, m.activeTab, m.width)

	return lipgloss.JoinVertical(lipgloss.Left, content, inputBox, statusBar)
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
	inputH := m.textarea.Height() + 2 // +2 for border
	fixed := inputH + 1               // input + statusBar
	vpHeight := m.height - fixed
	if vpHeight < 3 {
		vpHeight = 3
	}
	m.viewport.Width = m.width
	m.viewport.Height = vpHeight
	m.textarea.SetWidth(m.width - 6)
}

func (m *Model) updateViewport() {
	var stream string
	if m.streaming {
		stream = m.streamBuf
	}
	content := ui.RenderMessages(m.messages[m.activeTab], stream, m.viewport.Width)
	m.viewport.SetContent(content)
	m.viewport.GotoBottom()
}

func (m Model) currentModel() string {
	switch m.activeTab {
	case 1: // 개발
		return m.cfg.Models.Dev
	default: // 슈퍼택가이, 플랜 모두 super 모델 사용
		return m.cfg.Models.Super
	}
}

func (m *Model) sendMessage(input string) tea.Cmd {
	m.messages[m.activeTab] = append(m.messages[m.activeTab], ui.Message{
		Role: ui.RoleUser, Content: input, Timestamp: time.Now(),
	})
	m.histories[m.activeTab] = append(m.histories[m.activeTab], openai.ChatCompletionMessage{
		Role: openai.ChatMessageRoleUser, Content: input,
	})
	m.streaming = true
	m.streamBuf = ""
	m.tokenCount = 0
	m.streamStart = time.Now()
	m.updateViewport()

	ctx, cancel := context.WithCancel(context.Background())
	m.streamCancel = cancel
	model := m.currentModel()
	history := make([]openai.ChatCompletionMessage, len(m.histories[m.activeTab]))
	copy(history, m.histories[m.activeTab])
	m.streamCh = m.client.StreamChat(ctx, model, history)
	return m.waitForNextChunk()
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
		return streamChunkMsg{content: chunk.Content, done: chunk.Done, err: chunk.Err}
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
		m.messages[m.activeTab] = append(m.messages[m.activeTab], ui.Message{
			Role: ui.RoleAssistant, Content: m.streamBuf + "\n\n[중단됨]", Timestamp: time.Now(),
		})
		m.histories[m.activeTab] = append(m.histories[m.activeTab], openai.ChatCompletionMessage{
			Role: openai.ChatMessageRoleAssistant, Content: m.streamBuf,
		})
	}
	m.streamBuf = ""
	m.updateViewport()
}

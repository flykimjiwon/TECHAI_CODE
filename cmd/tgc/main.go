package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	tea "charm.land/bubbletea/v2"

	"github.com/kimjiwon/tgc/internal/app"
	"github.com/kimjiwon/tgc/internal/config"
	execpkg "github.com/kimjiwon/tgc/internal/exec"
	"github.com/kimjiwon/tgc/internal/llm"
)

func printDebugBanner(cfg config.Config) {
	fmt.Println()
	fmt.Println("  ╔══════════════════════════════════════════════╗")
	fmt.Println("  ║          [DEBUG MODE] 택가이코드             ║")
	fmt.Println("  ╚══════════════════════════════════════════════╝")
	fmt.Printf("  Version:   %s\n", version)
	fmt.Printf("  BaseURL:   %s\n", cfg.API.BaseURL)
	fmt.Printf("  Model:     %s\n", cfg.Models.Super)
	fmt.Printf("  ConfigDir: %s\n", config.ConfigDir())
	fmt.Printf("  LogFile:   %s\n", config.DebugLogPath())
	fmt.Println()
}

var version = "dev"

func main() {
	// Check for "exec" subcommand before flag parsing.
	// Scan all args (not just Args[1]) so "techai --debug exec" also works.
	for i, arg := range os.Args[1:] {
		if arg == "exec" {
			runExec(os.Args[i+2:])
			return
		}
		// Stop scanning at first non-flag argument that isn't "exec"
		if !strings.HasPrefix(arg, "-") {
			break
		}
	}

	modeFlag := flag.String("mode", "super", "시작 모드: super, dev, plan")
	multiFlag := flag.String("multi", "", "멀티 에이전트: on, off, review, consensus, scan, auto")
	versionFlag := flag.Bool("version", false, "버전 출력")
	setupFlag := flag.Bool("setup", false, "설정 재실행 (API URL/키 재입력)")
	resetFlag := flag.Bool("reset", false, "설정 초기화 (config 삭제 후 재설정)")
	flag.Parse()

	// Expose version to all packages so the TUI can display it.
	config.AppVersion = version

	if *versionFlag {
		fmt.Printf("택가이코드 (techai) %s\n", version)
		os.Exit(0)
	}

	// Handle --reset: delete config and force setup
	if *resetFlag {
		_ = os.Remove(config.ConfigPath())
		fmt.Println("  설정이 초기화되었습니다.")
		*setupFlag = true
	}

	// Migrate config models if needed (onprem: GPT-OSS → Qwen3-Coder)
	config.MigrateModelsIfNeeded(version)

	// Load config
	cfg, err := config.Load()
	if err != nil {
		cfg = config.DefaultConfig()
	}

	// Initialize debug logging (no-op if DebugMode != "true")
	config.InitDebugLog()
	defer config.CloseDebugLog()

	if config.IsDebug() {
		printDebugBanner(cfg)
		config.DebugLog("Config: baseURL=%s", cfg.API.BaseURL)
		config.DebugLog("Config: model=%s, configDir=%s", cfg.Models.Super, config.ConfigDir())
	}

	// Apply --multi flag override
	if *multiFlag != "" {
		switch *multiFlag {
		case "on", "true":
			cfg.Multi.Enabled = true
		case "off", "false":
			cfg.Multi.Enabled = false
		default:
			// Treat as strategy name
			cfg.Multi.Enabled = true
			cfg.Multi.Strategy = *multiFlag
		}
	}

	// Check if setup is needed (no API key) or forced via --setup
	needsSetup := config.NeedsSetup() || *setupFlag

	// Parse initial mode
	initialMode := parseMode(*modeFlag)

	// Create and run the app (AltScreen and Mouse are set in View)
	m := app.NewModel(cfg, initialMode, needsSetup)
	p := tea.NewProgram(m)

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "실행 오류: %v\n", err)
		os.Exit(1)
	}

	if config.IsDebug() {
		fmt.Printf("\n  [DEBUG] 로그 파일: %s\n\n", config.DebugLogPath())
	}
}

// runExec handles the "techai exec" headless subcommand.
func runExec(args []string) {
	execFlags := flag.NewFlagSet("exec", flag.ExitOnError)
	ephemeral := execFlags.Bool("ephemeral", false, "세션 저장 안 함 (일회성)")
	model := execFlags.String("model", "", "사용할 모델 (기본: config의 super 모델)")
	maxTurns := execFlags.Int("max-turns", 20, "최대 도구 실행 반복 횟수")
	execFlags.Parse(args)

	prompt := strings.Join(execFlags.Args(), " ")
	if prompt == "" {
		fmt.Fprintln(os.Stderr, "Usage: techai exec [--ephemeral] [--model MODEL] \"prompt\"")
		fmt.Fprintln(os.Stderr, "       echo input | techai exec \"prompt\"")
		os.Exit(1)
	}

	config.AppVersion = version

	cfg, err := config.Load()
	if err != nil {
		cfg = config.DefaultConfig()
	}

	config.InitDebugLog()
	defer config.CloseDebugLog()

	if cfg.API.APIKey == "" {
		fmt.Fprintln(os.Stderr, "Error: API key not configured. Run 'techai --setup' first.")
		os.Exit(1)
	}

	opts := execpkg.Options{
		Prompt:    prompt,
		Model:     *model,
		Mode:      llm.ModeSuper,
		MaxTurns:  *maxTurns,
		Ephemeral: *ephemeral,
	}

	// Read piped stdin
	stat, _ := os.Stdin.Stat()
	if (stat.Mode() & os.ModeCharDevice) == 0 {
		data := make([]byte, 0, 4096)
		buf := make([]byte, 4096)
		for {
			n, readErr := os.Stdin.Read(buf)
			if n > 0 {
				data = append(data, buf[:n]...)
			}
			if readErr != nil {
				break
			}
			if len(data) > 100000 {
				data = append(data[:100000], []byte("\n\n... [truncated]")...)
				break
			}
		}
		opts.Stdin = string(data)
	}

	if err := execpkg.Run(cfg, opts); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func parseMode(mode string) int {
	switch mode {
	case "super":
		return int(llm.ModeSuper)
	case "dev", "deep":
		return int(llm.ModeDev)
	case "plan":
		return int(llm.ModePlan)
	default:
		return int(llm.ModeSuper)
	}
}

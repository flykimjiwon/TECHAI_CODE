package main

import (
	"flag"
	"fmt"
	"os"

	tea "charm.land/bubbletea/v2"

	"github.com/kimjiwon/tgc/internal/app"
	"github.com/kimjiwon/tgc/internal/config"
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
	modeFlag := flag.String("mode", "super", "시작 모드: super, dev, plan")
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

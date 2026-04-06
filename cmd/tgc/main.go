package main

import (
	"flag"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/kimjiwon/tgc/internal/app"
	"github.com/kimjiwon/tgc/internal/config"
	"github.com/kimjiwon/tgc/internal/llm"
)

var version = "dev"

func main() {
	modeFlag := flag.String("mode", "super", "시작 모드: super, dev, plan")
	versionFlag := flag.Bool("version", false, "버전 출력")
	setupFlag := flag.Bool("setup", false, "설정 재실행 (API URL/키 재입력)")
	resetFlag := flag.Bool("reset", false, "설정 초기화 (config 삭제 후 재설정)")
	flag.Parse()

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

	// Check if setup is needed (no API key) or forced via --setup
	needsSetup := config.NeedsSetup() || *setupFlag

	// Parse initial mode
	initialMode := parseMode(*modeFlag)

	// Create and run the app (setup wizard runs inside TUI if needed)
	m := app.NewModel(cfg, initialMode, needsSetup)
	p := tea.NewProgram(m, tea.WithAltScreen(), tea.WithMouseCellMotion())

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "실행 오류: %v\n", err)
		os.Exit(1)
	}
}

func parseMode(mode string) int {
	switch mode {
	case "super", "슈퍼택가이":
		return int(llm.ModeSuper)
	case "dev", "개발":
		return int(llm.ModeDev)
	case "plan", "플랜":
		return int(llm.ModePlan)
	default:
		return int(llm.ModeSuper)
	}
}

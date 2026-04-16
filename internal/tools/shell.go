package tools

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"regexp"
	"runtime"
	"strings"
	"time"

	"github.com/kimjiwon/tgc/internal/config"
)

type ShellResult struct {
	Stdout   string
	Stderr   string
	ExitCode int
	Duration time.Duration
}

// Dangerous command patterns — block before execution.
var dangerousPatterns = []*regexp.Regexp{
	regexp.MustCompile(`\brm\s+(-[^\s]*\s+)*-[^\s]*r[^\s]*\s+/`),   // rm -rf /
	regexp.MustCompile(`\brm\s+(-[^\s]*\s+)*-[^\s]*r[^\s]*\s+~`),   // rm -rf ~
	regexp.MustCompile(`\bsudo\b`),                                   // sudo
	regexp.MustCompile(`\bmkfs\b`),                                   // mkfs
	regexp.MustCompile(`\bdd\s+.*of=/dev/`),                          // dd to device
	regexp.MustCompile(`>\s*/dev/sd`),                                // write to device
	regexp.MustCompile(`\bcurl\b.*\|\s*(sh|bash)`),                   // curl | bash
	regexp.MustCompile(`\bwget\b.*\|\s*(sh|bash)`),                   // wget | bash
	regexp.MustCompile(`:(){ :\|:& };:`),                             // fork bomb
	regexp.MustCompile(`\bshutdown\b`),                               // shutdown
	regexp.MustCompile(`\breboot\b`),                                 // reboot
	regexp.MustCompile(`\bchmod\s+777\s+/`),                          // chmod 777 /
}

// Risky command patterns — not blocked, but flagged with a warning in output.
var riskyPatterns = []struct {
	re   *regexp.Regexp
	desc string
}{
	{regexp.MustCompile(`\brm\s+-[^\s]*r`), "재귀 삭제 (rm -r)"},
	{regexp.MustCompile(`\bgit\s+reset\s+--hard`), "Git 하드 리셋"},
	{regexp.MustCompile(`\bgit\s+push\s+.*--force`), "Git 강제 푸시"},
	{regexp.MustCompile(`\bgit\s+push\s+.*-f\b`), "Git 강제 푸시"},
	{regexp.MustCompile(`\bgit\s+clean\s+.*-f`), "Git 미추적 파일 삭제"},
	{regexp.MustCompile(`\bgit\s+checkout\s+--\s*\.`), "Git 변경사항 폐기"},
	{regexp.MustCompile(`\bgit\s+branch\s+-D`), "Git 브랜치 강제 삭제"},
	{regexp.MustCompile(`\bdrop\s+(table|database)\b`), "DB 테이블/데이터베이스 삭제"},
	{regexp.MustCompile(`\btruncate\s+table\b`), "DB 테이블 비우기"},
	{regexp.MustCompile(`\bnpm\s+publish\b`), "npm 패키지 배포"},
	{regexp.MustCompile(`\bkill\s+-9`), "프로세스 강제 종료"},
}

// CheckSafety returns an error if the command matches a dangerous pattern.
// Also checks chained commands (&&, ;, |) individually to prevent bypass.
func CheckSafety(command string) error {
	// Split chained commands and check each segment
	segments := splitChainedCommands(command)
	for _, seg := range segments {
		lower := strings.ToLower(strings.TrimSpace(seg))
		for _, p := range dangerousPatterns {
			if p.MatchString(lower) {
				config.DebugLog("[SHELL-BLOCK] cmd=%q segment=%q matched pattern=%s", command, seg, p.String())
				return fmt.Errorf("blocked: dangerous command pattern detected in: %s", strings.TrimSpace(seg))
			}
		}
	}
	return nil
}

// splitChainedCommands splits a command string on &&, ||, ;, and pipe operators
// to check each segment independently against safety patterns.
func splitChainedCommands(command string) []string {
	// Replace chain operators with a unique separator
	for _, sep := range []string{"&&", "||", ";", "|"} {
		command = strings.ReplaceAll(command, sep, "\x00")
	}
	// Also handle $() and backtick subshells
	command = strings.ReplaceAll(command, "$(", "\x00")
	command = strings.ReplaceAll(command, "`", "\x00")

	parts := strings.Split(command, "\x00")
	var result []string
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			result = append(result, p)
		}
	}
	return result
}

// CheckRisky returns a warning string if the command is risky but not blocked.
// Returns empty string if the command is safe.
func CheckRisky(command string) string {
	lower := strings.ToLower(command)
	for _, p := range riskyPatterns {
		if p.re.MatchString(lower) {
			config.DebugLog("[SHELL-WARN] cmd=%q risky: %s", command, p.desc)
			return fmt.Sprintf("⚠ 주의: %s", p.desc)
		}
	}
	return ""
}

func ShellExec(ctx context.Context, command string) (ShellResult, error) {
	// Safety check
	if err := CheckSafety(command); err != nil {
		return ShellResult{ExitCode: -1}, err
	}

	cwd := ""
	if dir, err := exec.LookPath("sh"); err == nil {
		cwd = dir
	}
	deadline, hasDeadline := ctx.Deadline()
	timeout := "none"
	if hasDeadline {
		timeout = fmt.Sprintf("%v", time.Until(deadline).Round(time.Millisecond))
	}
	config.DebugLog("[SHELL] cmd=%q | cwd=%s | timeout=%s | os=%s", command, cwd, timeout, runtime.GOOS)

	start := time.Now()

	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.CommandContext(ctx, "cmd", "/c", command)
	} else {
		cmd = exec.CommandContext(ctx, "sh", "-c", command)
	}

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	duration := time.Since(start)

	result := ShellResult{
		Stdout:   stdout.String(),
		Stderr:   stderr.String(),
		ExitCode: 0,
		Duration: duration,
	}

	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			result.ExitCode = exitErr.ExitCode()
		} else {
			config.DebugLog("[SHELL-ERR] cmd=%q | err=%v | elapsed=%v", command, err, duration)
			return result, fmt.Errorf("exec failed: %w", err)
		}
	}

	config.DebugLog("[SHELL] exitCode=%d | stdout=%dbytes | stderr=%dbytes | elapsed=%v", result.ExitCode, len(result.Stdout), len(result.Stderr), duration)
	return result, nil
}

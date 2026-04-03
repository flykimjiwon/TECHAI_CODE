package tools

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"regexp"
	"strings"
	"time"
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

// CheckSafety returns an error if the command matches a dangerous pattern.
func CheckSafety(command string) error {
	lower := strings.ToLower(command)
	for _, p := range dangerousPatterns {
		if p.MatchString(lower) {
			return fmt.Errorf("blocked: dangerous command pattern detected: %s", p.String())
		}
	}
	return nil
}

func ShellExec(ctx context.Context, command string) (ShellResult, error) {
	// Safety check
	if err := CheckSafety(command); err != nil {
		return ShellResult{ExitCode: -1}, err
	}

	start := time.Now()

	cmd := exec.CommandContext(ctx, "sh", "-c", command)

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
			return result, fmt.Errorf("exec failed: %w", err)
		}
	}

	return result, nil
}

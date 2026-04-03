package tools

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"time"
)

type ShellResult struct {
	Stdout   string
	Stderr   string
	ExitCode int
	Duration time.Duration
}

func ShellExec(ctx context.Context, command string) (ShellResult, error) {
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

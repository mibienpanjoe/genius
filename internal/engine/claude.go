package engine

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"
)

// claudeEngine drives the `claude` CLI in print mode. Output is plain text
// (default --output-format text), so the final answer is stdout, trimmed.
type claudeEngine struct {
	model string
}

func (e *claudeEngine) Name() string { return "claude" }

func (e *claudeEngine) Generate(ctx context.Context, sys, user string) (string, error) {
	if _, err := exec.LookPath("claude"); err != nil {
		return "", fmt.Errorf("%w: claude", ErrNotInstalled)
	}

	args := []string{"-p", user}
	if strings.TrimSpace(sys) != "" {
		args = append(args, "--append-system-prompt", sys)
	}
	if e.model != "" {
		args = append(args, "--model", e.model)
	}

	cmd := exec.CommandContext(ctx, "claude", args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("claude failed: %w: %s", err, strings.TrimSpace(stderr.String()))
	}
	return strings.TrimSpace(stdout.String()), nil
}

// Describe is unsupported on the claude print path (no confirmed image-input
// flag, ADR-06); callers fall back to placeholders (EXC-12).
func (e *claudeEngine) Describe(_ context.Context, _, _ string) (string, error) {
	return "", ErrNoVision
}

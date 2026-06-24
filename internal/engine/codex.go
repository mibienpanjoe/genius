package engine

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// codexEngine drives `codex exec`. The system and user prompts are combined and
// passed on stdin; the final assistant message is captured via
// `-o <file>` (--output-last-message), which yields clean output free of the
// agent's event framing (INV-07).
type codexEngine struct {
	model string
}

func (e *codexEngine) Name() string { return "codex" }

func (e *codexEngine) Generate(ctx context.Context, sys, user string) (string, error) {
	if _, err := exec.LookPath("codex"); err != nil {
		return "", fmt.Errorf("%w: codex", ErrNotInstalled)
	}

	outFile, err := os.CreateTemp("", "genius-codex-*.txt")
	if err != nil {
		return "", fmt.Errorf("codex: temp file: %w", err)
	}
	outFile.Close()
	defer os.Remove(outFile.Name())

	args := []string{"exec", "--color", "never", "-o", outFile.Name()}
	if e.model != "" {
		args = append(args, "-m", e.model)
	}
	args = append(args, "-") // read prompt from stdin

	prompt := user
	if strings.TrimSpace(sys) != "" {
		prompt = sys + "\n\n" + user
	}

	cmd := exec.CommandContext(ctx, "codex", args...)
	cmd.Stdin = strings.NewReader(prompt)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("codex failed: %w: %s", err, strings.TrimSpace(stderr.String()))
	}

	data, err := os.ReadFile(outFile.Name())
	if err != nil {
		return "", fmt.Errorf("codex: reading output: %w", err)
	}
	return strings.TrimSpace(string(data)), nil
}

// Describe attaches an image via `codex exec -i` and returns the transcription,
// captured cleanly through -o (FR-035b/f).
func (e *codexEngine) Describe(ctx context.Context, imagePath, instruction string) (string, error) {
	if _, err := exec.LookPath("codex"); err != nil {
		return "", fmt.Errorf("%w: codex", ErrNotInstalled)
	}
	outFile, err := os.CreateTemp("", "genius-codex-img-*.txt")
	if err != nil {
		return "", fmt.Errorf("codex: temp file: %w", err)
	}
	outFile.Close()
	defer os.Remove(outFile.Name())

	// -i is variadic (<FILE>...), so it would swallow a positional prompt;
	// pass the instruction via stdin instead.
	args := []string{"exec", "--color", "never", "-o", outFile.Name()}
	if e.model != "" {
		args = append(args, "-m", e.model)
	}
	args = append(args, "-i", imagePath)

	cmd := exec.CommandContext(ctx, "codex", args...)
	cmd.Stdin = strings.NewReader(instruction)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("codex describe failed: %w: %s", err, strings.TrimSpace(stderr.String()))
	}
	data, err := os.ReadFile(outFile.Name())
	if err != nil {
		return "", fmt.Errorf("codex: reading output: %w", err)
	}
	return strings.TrimSpace(string(data)), nil
}

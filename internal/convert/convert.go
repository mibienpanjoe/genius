// Package convert turns source documents (PDF/PPT/…) into markdown by invoking
// the markitdown CLI (the Converter actor). Phase 3 handles the text layer;
// image extraction and notation-fidelity transcription arrive in Phase 6.
package convert

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strings"
)

// ErrNotInstalled signals markitdown is not on PATH (ERR-031, INV-09).
var ErrNotInstalled = errors.New("markitdown not found")

// InstallHint is the actionable remedy shown when the Converter is absent.
const InstallHint = "markitdown not found — run: pip install markitdown"

// Available reports whether the markitdown CLI is on PATH.
func Available() bool {
	_, err := exec.LookPath("markitdown")
	return err == nil
}

// ToMarkdown converts the document at path to markdown via markitdown and
// returns the markdown text. A missing converter yields ErrNotInstalled.
func ToMarkdown(ctx context.Context, path string) (string, error) {
	if !Available() {
		return "", ErrNotInstalled
	}
	cmd := exec.CommandContext(ctx, "markitdown", path)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("markitdown failed on %s: %w: %s",
			path, err, strings.TrimSpace(stderr.String()))
	}
	md := strings.TrimSpace(stdout.String())
	if md == "" {
		return "", fmt.Errorf("markitdown produced no text for %s "+
			"(scanned/image-only? --ocr support is planned)", path)
	}
	return md, nil
}

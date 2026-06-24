// Package engine abstracts the generation backend behind a single interface so
// claude and codex are swappable without changing the workflow (INV-06).
package engine

import (
	"context"
	"errors"
	"fmt"
)

// Engine is a generation backend (docs/06 §Go interfaces). All generation
// flows through this interface (INV-06, FR-081).
type Engine interface {
	// Name reports the backend identifier ("claude" | "codex").
	Name() string
	// Generate runs one non-interactive turn with a system prompt and a user
	// prompt and returns only the final assistant text (INV-07, FR-084).
	Generate(ctx context.Context, sys, user string) (string, error)
	// Describe transcribes/describes an image given an instruction, for figure
	// captioning and notation transcription during ingest (FR-035b/f). Backends
	// without image support return ErrNoVision so the caller can fall back to a
	// placeholder (EXC-12).
	Describe(ctx context.Context, imagePath, instruction string) (string, error)
}

// ErrNotInstalled signals the backend binary is not on PATH (ERR-081, INV-09).
var ErrNotInstalled = errors.New("engine binary not found on PATH")

// ErrNoVision signals the backend cannot accept image input; callers fall back
// to an asset-referencing placeholder (ERR-036/EXC-12).
var ErrNoVision = errors.New("engine has no vision support")

// New constructs an Engine by name with an optional model override. Unknown
// names are an error.
func New(name, model string) (Engine, error) {
	switch name {
	case "", "claude":
		return &claudeEngine{model: model}, nil
	case "codex":
		return &codexEngine{model: model}, nil
	default:
		return nil, fmt.Errorf("unknown engine %q (want claude|codex)", name)
	}
}

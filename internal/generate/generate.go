// Package generate assembles prompts and drives the Engine to produce study
// guides and revision Q&A grounded in course material (INV-04).
package generate

import (
	"context"
	_ "embed"
	"fmt"
	"strings"

	"genius/internal/engine"
)

//go:embed prompts/guide.txt
var guidePrompt string

//go:embed prompts/qa.txt
var qaPrompt string

// DefaultQACount is applied when the learner does not specify a count (FR-054).
const DefaultQACount = 10

// Guide produces a study guide from the course material (FR-040..042).
func Guide(ctx context.Context, eng engine.Engine, course, material string) (string, error) {
	if strings.TrimSpace(material) == "" {
		return "", fmt.Errorf("no material to ground on")
	}
	user := fmt.Sprintf("Course: %s\n\nCourse material:\n\n%s\n\nWrite the study guide now.",
		course, material)
	return eng.Generate(ctx, guidePrompt, user)
}

// QAOpts parameterizes Q&A generation (FR-054/055).
type QAOpts struct {
	Count int    // target number of pairs; <=0 means DefaultQACount
	Scope string // optional free-text focus constraint
}

// QA produces revision Q&A from the course material in the learner's format
// (FR-050..056).
func QA(ctx context.Context, eng engine.Engine, course, material string, opts QAOpts) (string, error) {
	if strings.TrimSpace(material) == "" {
		return "", fmt.Errorf("no material to ground on")
	}
	count := opts.Count
	if count <= 0 {
		count = DefaultQACount
	}

	var b strings.Builder
	fmt.Fprintf(&b, "Course: %s\n\n", course)
	fmt.Fprintf(&b, "Generate exactly %d Q&A pairs.\n", count)
	if s := strings.TrimSpace(opts.Scope); s != "" {
		fmt.Fprintf(&b, "Scope — draw ONLY from this part of the course: %s\n", s)
	}
	fmt.Fprintf(&b, "\nCourse material:\n\n%s\n\nWrite the Q&A now, starting at `## Q1.`.", material)

	return eng.Generate(ctx, qaPrompt, b.String())
}

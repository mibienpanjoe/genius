package cli

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mibienpanjoe/genius/internal/convert"
)

const solveSampleSet = `# TD 1

Exercice 1
Simplify F = AB + AB̄.

## Exercice 2
1. Truth table.
2. Karnaugh map.
`

// runCLI executes the root command with args, capturing os.Stdout (the
// subcommands print there directly).
func runCLI(t *testing.T, args ...string) (string, error) {
	t.Helper()
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	cmd := NewRootCmd()
	cmd.SetArgs(args)
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)
	err := cmd.Execute()

	_ = w.Close()
	os.Stdout = old
	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)
	return buf.String(), err
}

func seedSet(t *testing.T, dir, course, set, body string) {
	t.Helper()
	p := filepath.Join(dir, "exercises", course, set+".md")
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestSolveListMode(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("GENIUS_HOME", dir)
	seedSet(t, dir, "algebra", "td1", solveSampleSet)

	out, err := runCLI(t, "solve", "algebra", "--set", "td1")
	if err != nil {
		t.Fatalf("list mode errored: %v", err)
	}
	if !strings.Contains(out, "Exercice 1") || !strings.Contains(out, "Exercice 2") {
		t.Errorf("exercises not listed: %q", out)
	}
	if !strings.Contains(out, "--ex") {
		t.Errorf("missing --ex hint: %q", out)
	}
	// Sub-parts of exercise 2 are addressable.
	if !strings.Contains(out, "2.1") || !strings.Contains(out, "2.2") {
		t.Errorf("sub-parts not listed: %q", out)
	}
}

func TestSolveUnknownSet(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("GENIUS_HOME", dir)
	seedSet(t, dir, "algebra", "td1", solveSampleSet)

	_, err := runCLI(t, "solve", "algebra", "--set", "nope")
	if err == nil || !strings.Contains(err.Error(), "available: td1") {
		t.Errorf("want unknown-set error listing td1, got %v", err)
	}
}

func TestSolveNoSetsAtAll(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("GENIUS_HOME", dir)

	_, err := runCLI(t, "solve", "history", "--set", "td1")
	if err == nil || !strings.Contains(err.Error(), "no exercise sets") {
		t.Errorf("want no-sets guidance, got %v", err)
	}
}

// An exercise ingest aimed at a course that doesn't exist must refuse before
// converting — otherwise the set files under a name no dashboard lists.
func TestIngestExerciseUnknownCourse(t *testing.T) {
	if !convert.Available() {
		t.Skip("markitdown not on PATH")
	}
	dir := t.TempDir()
	t.Setenv("GENIUS_HOME", dir)
	src := filepath.Join(dir, "td1.md")
	if err := os.WriteFile(src, []byte("Exercice 1\nx\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	_, err := runCLI(t, "ingest", src, "--kind", "exercise", "--course", "ghost")
	if err == nil || !strings.Contains(err.Error(), `no course "ghost"`) {
		t.Errorf("want unknown-course refusal, got %v", err)
	}
	if _, serr := os.Stat(filepath.Join(dir, "exercises", "ghost")); serr == nil {
		t.Error("refused ingest must not create the orphaned exercises dir")
	}
}

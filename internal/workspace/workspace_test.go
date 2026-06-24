package workspace

import (
	"os"
	"path/filepath"
	"testing"
)

func TestResolveRootPrecedence(t *testing.T) {
	t.Setenv("GENIUS_HOME", "/tmp/forced-root")
	if got := ResolveRoot(Config{StudyRoot: "/ignored"}); got != "/tmp/forced-root" {
		t.Fatalf("env should win, got %s", got)
	}

	t.Setenv("GENIUS_HOME", "")
	if got := ResolveRoot(Config{StudyRoot: "/cfg/root"}); got != "/cfg/root" {
		t.Fatalf("config should win when no env, got %s", got)
	}
}

func TestOpenCreatesTree(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("GENIUS_HOME", dir)

	w, err := Open(Config{})
	if err != nil {
		t.Fatal(err)
	}
	for _, d := range subdirs {
		if _, err := os.Stat(filepath.Join(w.Root, d)); err != nil {
			t.Errorf("subdir %s not created: %v", d, err)
		}
	}
}

func TestCoursesScanCounts(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("GENIUS_HOME", dir)
	w, err := Open(Config{})
	if err != nil {
		t.Fatal(err)
	}

	// algebra: guide + qa + 2 exercise sets
	mustMkdir(t, w.Path("courses", "algebra"))
	mustWrite(t, w.Path("courses", "algebra", "lecture.md"), "# x")
	mustWrite(t, w.Path("guides", "algebra.md"), "# guide")
	mustWrite(t, w.Path("qa", "algebra.md"), "## Q1. x")
	mustMkdir(t, w.Path("exercises", "algebra"))
	mustWrite(t, w.Path("exercises", "algebra", "td1.md"), "# td1")
	mustWrite(t, w.Path("exercises", "algebra", "td2.md"), "# td2")

	// history: nothing derived
	mustMkdir(t, w.Path("courses", "history"))

	courses, err := w.Courses()
	if err != nil {
		t.Fatal(err)
	}
	if len(courses) != 2 {
		t.Fatalf("want 2 courses, got %d", len(courses))
	}
	// sorted: algebra, history
	a := courses[0]
	if a.Name != "algebra" || !a.HasGuide || !a.HasQA || a.ExerciseSets != 2 {
		t.Errorf("algebra counts wrong: %+v", a)
	}
	h := courses[1]
	if h.Name != "history" || h.HasGuide || h.HasQA || h.ExerciseSets != 0 {
		t.Errorf("history counts wrong: %+v", h)
	}
}

func mustMkdir(t *testing.T, p string) {
	t.Helper()
	if err := os.MkdirAll(p, 0o755); err != nil {
		t.Fatal(err)
	}
}

func mustWrite(t *testing.T, p, body string) {
	t.Helper()
	if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

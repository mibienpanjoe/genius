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

func TestChapterArtifactPaths(t *testing.T) {
	w := Workspace{Root: "/study"}
	if got := w.ChapterGuidePath("algebra", "chap01"); got != "/study/guides/algebra/chap01.md" {
		t.Errorf("ChapterGuidePath wrong: %s", got)
	}
	if got := w.ChapterQAPath("algebra", "chap01+chap02"); got != "/study/qa/algebra/chap01+chap02.md" {
		t.Errorf("ChapterQAPath wrong: %s", got)
	}
}

func TestCoursesCountsChapterArtifacts(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("GENIUS_HOME", dir)
	w, err := Open(Config{})
	if err != nil {
		t.Fatal(err)
	}

	mustMkdir(t, w.Path("courses", "algebra"))
	mustMkdir(t, w.Path("guides", "algebra"))
	mustMkdir(t, w.Path("qa", "algebra"))
	mustWrite(t, w.Path("guides", "algebra.md"), "# whole guide")        // whole
	mustWrite(t, w.Path("guides", "algebra", "chap01.md"), "# c1")       // scoped
	mustWrite(t, w.Path("guides", "algebra", "chap01+chap02.md"), "# s") // scoped span
	mustWrite(t, w.Path("qa", "algebra", "chap02.md"), "## Q1. x")       // scoped qa, no whole qa

	courses, err := w.Courses()
	if err != nil {
		t.Fatal(err)
	}
	c := courses[0]
	// guide: whole(1) + 2 scoped = 3
	if c.GuideCount() != 3 {
		t.Errorf("GuideCount want 3, got %d (%+v)", c.GuideCount(), c)
	}
	// qa: no whole + 1 scoped = 1
	if c.QACount() != 1 {
		t.Errorf("QACount want 1, got %d (%+v)", c.QACount(), c)
	}
	if c.HasGuide != true || c.HasQA != false {
		t.Errorf("whole flags wrong: HasGuide=%v HasQA=%v", c.HasGuide, c.HasQA)
	}

	gs, _ := w.GuideScopes("algebra")
	if len(gs) != 2 || gs[0] != "chap01" || gs[1] != "chap01+chap02" {
		t.Errorf("GuideScopes wrong: %v", gs)
	}
	qs, _ := w.QAScopes("algebra")
	if len(qs) != 1 || qs[0] != "chap02" {
		t.Errorf("QAScopes wrong: %v", qs)
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

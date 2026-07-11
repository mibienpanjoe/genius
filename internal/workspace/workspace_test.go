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

func TestScopeName(t *testing.T) {
	if got := ScopeName([]string{"chap01.md"}); got != "chap01" {
		t.Errorf("single scope want chap01, got %q", got)
	}
	// unsorted input is sorted and joined
	if got := ScopeName([]string{"chap02.md", "chap01.md"}); got != "chap01+chap02" {
		t.Errorf("combined scope want chap01+chap02, got %q", got)
	}
	// duplicates collapse to one slug
	if got := ScopeName([]string{"chap01.md", "chap01.md"}); got != "chap01" {
		t.Errorf("duplicate files should dedupe, got %q", got)
	}
}

func TestGuideQATarget(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("GENIUS_HOME", dir)
	w, err := Open(Config{})
	if err != nil {
		t.Fatal(err)
	}
	mustMkdir(t, w.Path("courses", "algebra"))
	for _, f := range []string{"chap01.md", "chap02.md", "chap03.md"} {
		mustWrite(t, w.Path("courses", "algebra", f), "# x")
	}

	// no files / all files → whole-course slot
	if got := w.GuideTarget("algebra", nil); got != w.GuidePath("algebra") {
		t.Errorf("nil files should target the whole guide, got %s", got)
	}
	all := []string{"chap01.md", "chap02.md", "chap03.md"}
	if got := w.QATarget("algebra", all); got != w.QAPath("algebra") {
		t.Errorf("all-selected should target the whole qa, got %s", got)
	}
	// subset → scoped artifact
	if got := w.GuideTarget("algebra", []string{"chap01.md"}); got != w.ChapterGuidePath("algebra", "chap01") {
		t.Errorf("single chapter should target a scoped guide, got %s", got)
	}
	if got := w.QATarget("algebra", []string{"chap01.md", "chap02.md"}); got != w.ChapterQAPath("algebra", "chap01+chap02") {
		t.Errorf("span should target a combined qa, got %s", got)
	}
	// duplicated file matching the chapter count must NOT claim the whole-course
	// slot — it is still a single-chapter scope.
	dup := []string{"chap01.md", "chap01.md", "chap01.md"}
	if got := w.GuideTarget("algebra", dup); got != w.ChapterGuidePath("algebra", "chap01") {
		t.Errorf("duplicated selection should stay scoped, got %s", got)
	}
}

// solve --save output (<set>.solutions.md) must not surface as a solvable
// exercise set nor inflate the dashboard chip count.
func TestSolutionsFileNotAnExerciseSet(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("GENIUS_HOME", dir)
	w, err := Open(Config{})
	if err != nil {
		t.Fatal(err)
	}
	mustMkdir(t, w.Path("courses", "algebra"))
	mustMkdir(t, w.Path("exercises", "algebra"))
	mustWrite(t, w.Path("exercises", "algebra", "td1.md"), "# td1")
	mustWrite(t, w.Path("exercises", "algebra", "td1.solutions.md"), "# solved")

	sets, err := w.ExerciseSets("algebra")
	if err != nil {
		t.Fatal(err)
	}
	if len(sets) != 1 || sets[0] != "td1" {
		t.Errorf("solutions file must be excluded from sets, got %v", sets)
	}

	courses, err := w.Courses()
	if err != nil {
		t.Fatal(err)
	}
	if courses[0].ExerciseSets != 1 {
		t.Errorf("solutions file must not count as a set, got %d", courses[0].ExerciseSets)
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

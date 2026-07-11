package tui

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/mibienpanjoe/genius/internal/engine"
	"github.com/mibienpanjoe/genius/internal/quiz"
	"github.com/mibienpanjoe/genius/internal/workspace"
)

func keyRunes(s string) tea.KeyMsg { return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(s)} }

func TestHomeRenders(t *testing.T) {
	m := New("claude", nil, workspace.Workspace{Root: "/home/u/study"}, nil)
	mm, _ := m.Update(tea.WindowSizeMsg{Width: 100, Height: 30})
	out := mm.View()

	for _, want := range []string{"getting started:", "No courses yet", "engine:claude", "/home/u/study"} {
		if !strings.Contains(out, want) {
			t.Errorf("home view missing %q", want)
		}
	}
}

func TestQuitKey(t *testing.T) {
	m := New("claude", nil, workspace.Workspace{Root: "/x"}, nil)
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyCtrlC})
	if cmd == nil {
		t.Error("ctrl+c should quit from home")
	}
}

func TestPopulatedHomeAndNav(t *testing.T) {
	courses := []workspace.Course{
		{Name: "algebra", HasGuide: true, HasQA: true, ExerciseSets: 1},
		{Name: "history", ExerciseSets: 0},
	}
	m := New("codex", nil, workspace.Workspace{Root: "/s"}, courses)
	mm, _ := m.Update(tea.WindowSizeMsg{Width: 100, Height: 30})
	out := mm.View()
	for _, want := range []string{"COURSES", "algebra", "history", "g·1", "e·0"} {
		if !strings.Contains(out, want) {
			t.Errorf("populated home missing %q", want)
		}
	}

	// cursor starts at 0; down moves to 1, clamps there.
	model := mm.(Model)
	if model.cursor != 0 {
		t.Fatalf("cursor should start 0, got %d", model.cursor)
	}
	down := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("j")}
	m2, _ := model.Update(down)
	if m2.(Model).cursor != 1 {
		t.Errorf("down should move cursor to 1")
	}
	m3, _ := m2.(Model).Update(down)
	if m3.(Model).cursor != 1 {
		t.Errorf("cursor should clamp at last index")
	}
}

func TestSolveFlowNavigation(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("GENIUS_HOME", dir)
	ws, err := workspace.Open(workspace.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(ws.Path("exercises", "algebra"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(ws.Path("exercises", "algebra", "td1.md"),
		[]byte("Exercice 1\nFoo.\n\nExercice 2\nBar.\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	courses := []workspace.Course{{Name: "algebra", ExerciseSets: 1}}

	m := New("claude", nil, ws, courses)
	mm, _ := m.Update(tea.WindowSizeMsg{Width: 100, Height: 30})

	// s → set picker, lists the set.
	s, _ := mm.(Model).Update(keyRunes("s"))
	if s.(Model).state != stateExSets {
		t.Fatalf("s should open set picker, state=%d", s.(Model).state)
	}
	if !strings.Contains(s.View(), "td1") {
		t.Errorf("set not listed: %q", s.View())
	}

	// enter → exercise list, enumerated.
	el, _ := s.(Model).Update(tea.KeyMsg{Type: tea.KeyEnter})
	if el.(Model).state != stateExList {
		t.Fatalf("enter should open exercise list, state=%d", el.(Model).state)
	}
	out := el.View()
	if !strings.Contains(out, "Exercice 1") || !strings.Contains(out, "Exercice 2") {
		t.Errorf("exercises not listed: %q", out)
	}

	// space toggles selection on the current item.
	sel, _ := el.(Model).Update(keyRunes(" "))
	if !sel.(Model).exSelected[0] {
		t.Errorf("space should select the current exercise")
	}

	// enter with a nil engine bounces home with a notice (engine guard).
	done, _ := sel.(Model).Update(tea.KeyMsg{Type: tea.KeyEnter})
	if done.(Model).state != stateHome {
		t.Errorf("nil engine should return home, state=%d", done.(Model).state)
	}
	if !strings.Contains(done.(Model).notice, "no engine") {
		t.Errorf("want no-engine notice, got %q", done.(Model).notice)
	}
}

func TestGenerateGuideFlow(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("GENIUS_HOME", dir)
	ws, err := workspace.Open(workspace.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(ws.Path("courses", "algebra"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(ws.Path("courses", "algebra", "chap01.md"),
		[]byte("# Algebra\nRings and fields.\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	eng := &engine.Fake{Reply: "# Study Guide\n\nKey ideas."}
	courses := []workspace.Course{{Name: "algebra"}} // no guide yet

	m := New("fake", eng, ws, courses)
	mm, _ := m.Update(tea.WindowSizeMsg{Width: 100, Height: 30})

	// g with no guide present → enters the generating state (spinner).
	g, _ := mm.(Model).Update(keyRunes("g"))
	if g.(Model).state != stateGenerating {
		t.Fatalf("g should enter generating, state=%d", g.(Model).state)
	}

	// Drive the async generate command and feed its result back.
	gm := g.(Model)
	done, _ := gm.Update(genCmd(context.Background(), eng, "guide", "algebra", "material", 0, gm.workEpoch)())
	dm := done.(Model)
	if eng.Calls != 1 {
		t.Errorf("engine should be called once, got %d", eng.Calls)
	}
	if dm.state != stateReader {
		t.Fatalf("generate should land in the reader, state=%d", dm.state)
	}
	if !dm.courses[0].HasGuide {
		t.Errorf("chip count should flip: HasGuide=false")
	}
	data, err := os.ReadFile(ws.GuidePath("algebra"))
	if err != nil {
		t.Fatalf("guide not written: %v", err)
	}
	if !strings.Contains(string(data), "Study Guide") {
		t.Errorf("written guide missing content: %q", data)
	}
}

func TestGenerateNoMaterialRefuses(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("GENIUS_HOME", dir)
	ws, err := workspace.Open(workspace.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(ws.Path("courses", "empty"), 0o755); err != nil {
		t.Fatal(err)
	}
	eng := &engine.Fake{Reply: "should not be called"}
	courses := []workspace.Course{{Name: "empty"}}

	m := New("fake", eng, ws, courses)
	mm, _ := m.Update(tea.WindowSizeMsg{Width: 100, Height: 30})

	g, _ := mm.(Model).Update(keyRunes("g"))
	gm := g.(Model)
	if gm.state != stateHome {
		t.Errorf("no material should stay home, state=%d", gm.state)
	}
	if eng.Calls != 0 {
		t.Errorf("engine must not run without material, got %d calls", eng.Calls)
	}
	if !strings.Contains(gm.notice, "no source material") {
		t.Errorf("want no-material notice, got %q", gm.notice)
	}
}

// TestScopedQAFlow exercises the chapter-scoped Q&A path end to end: the count
// prompt gates generation, the spinner names the chosen chapter scope (not just
// the course), and closing the reader returns to the chapter hub it came from.
func TestScopedQAFlow(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("GENIUS_HOME", dir)
	ws, err := workspace.Open(workspace.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(ws.Path("courses", "algebra"), 0o755); err != nil {
		t.Fatal(err)
	}
	for _, f := range []string{"chap01.md", "chap02.md"} {
		if err := os.WriteFile(ws.Path("courses", "algebra", f),
			[]byte("# Algebra\nRings and fields.\n"), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	eng := &engine.Fake{Reply: "## Q1. What is a ring?\n\nA set with two operations."}
	courses := []workspace.Course{{Name: "algebra"}}

	m := New("fake", eng, ws, courses)
	mm, _ := m.Update(tea.WindowSizeMsg{Width: 100, Height: 30})

	// Enter the chapter hub and scope to a single chapter.
	fm, _ := mm.(Model).Update(keyRunes("f"))
	if fm.(Model).state != stateChapters {
		t.Fatalf("f should open chapters, state=%d", fm.(Model).state)
	}
	scoped := workspace.Slug(fm.(Model).chapFiles[0])
	sel, _ := fm.(Model).Update(keyRunes(" ")) // select chapter at cursor 0

	// Q&A (force) → the count prompt, remembering chapters as the return screen.
	pm, _ := sel.(Model).Update(keyRunes("Q"))
	pv := pm.(Model)
	if pv.state != stateQACount {
		t.Fatalf("Q should prompt for count, state=%d", pv.state)
	}
	if pv.back != stateChapters {
		t.Errorf("count prompt should return to chapters, back=%d", pv.back)
	}

	// Type a count and confirm.
	typed, _ := pv.Update(keyRunes("8"))
	gen, _ := typed.(Model).Update(tea.KeyMsg{Type: tea.KeyEnter})
	gv := gen.(Model)
	if gv.state != stateGenerating {
		t.Fatalf("enter should start generating, state=%d", gv.state)
	}
	if gv.genCount != 8 {
		t.Errorf("genCount should be 8, got %d", gv.genCount)
	}
	wantScope := "algebra/" + scoped
	if gv.genScope != wantScope {
		t.Errorf("genScope = %q, want %q", gv.genScope, wantScope)
	}
	if out := gv.viewGenerating(); !strings.Contains(out, wantScope) {
		t.Errorf("spinner should name the chapter scope %q, got %q", wantScope, out)
	}

	// Drive the async generate and land in the reader.
	done, _ := gv.Update(genCmd(context.Background(), eng, "qa", "algebra", "material", 8, gv.workEpoch)())
	dm := done.(Model)
	if dm.state != stateReader {
		t.Fatalf("generate should open the reader, state=%d", dm.state)
	}
	if _, err := os.Stat(ws.ChapterQAPath("algebra", scoped)); err != nil {
		t.Errorf("scoped Q&A not written: %v", err)
	}

	// Closing the reader returns to the chapter hub, not home.
	back, _ := dm.Update(keyRunes("q"))
	if back.(Model).state != stateChapters {
		t.Errorf("reader back should return to chapters, state=%d", back.(Model).state)
	}
}

// TestQACountDefault confirms a blank count leaves genCount at 0 so the
// generator applies DefaultQACount.
func TestQACountDefault(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("GENIUS_HOME", dir)
	ws, err := workspace.Open(workspace.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(ws.Path("courses", "algebra"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(ws.Path("courses", "algebra", "chap01.md"),
		[]byte("# Algebra\nRings.\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	eng := &engine.Fake{Reply: "## Q1. x\n\ny"}
	m := New("fake", eng, ws, []workspace.Course{{Name: "algebra"}})
	mm, _ := m.Update(tea.WindowSizeMsg{Width: 100, Height: 30})

	// q on a course with no Q&A → count prompt, whole-course scope.
	pm, _ := mm.(Model).Update(keyRunes("q"))
	pv := pm.(Model)
	if pv.state != stateQACount {
		t.Fatalf("q should prompt for count, state=%d", pv.state)
	}
	if pv.genScope != "" && pv.back != stateHome {
		t.Errorf("home Q&A should return home, back=%d", pv.back)
	}
	// Enter with no input → default (genCount stays 0).
	gen, _ := pv.Update(tea.KeyMsg{Type: tea.KeyEnter})
	gv := gen.(Model)
	if gv.state != stateGenerating {
		t.Fatalf("enter should start generating, state=%d", gv.state)
	}
	if gv.genCount != 0 {
		t.Errorf("blank count should leave genCount 0 (default), got %d", gv.genCount)
	}
	if gv.genScope != "algebra" {
		t.Errorf("whole-course scope should be %q, got %q", "algebra", gv.genScope)
	}
}

// TestCancelGenerationStaysInApp verifies that stopping a generation returns to
// the launching screen instead of quitting genius, and that the abandoned
// result is discarded when it finally arrives.
func TestCancelGenerationStaysInApp(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("GENIUS_HOME", dir)
	ws, err := workspace.Open(workspace.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(ws.Path("courses", "algebra"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(ws.Path("courses", "algebra", "chap01.md"),
		[]byte("# Algebra\nRings.\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	eng := &engine.Fake{Reply: "# Study Guide\n\nKey ideas."}
	m := New("fake", eng, ws, []workspace.Course{{Name: "algebra"}})
	mm, _ := m.Update(tea.WindowSizeMsg{Width: 100, Height: 30})

	// Start a guide generation, then abort it with esc.
	g, _ := mm.(Model).Update(keyRunes("g"))
	gv := g.(Model)
	if gv.state != stateGenerating {
		t.Fatalf("g should enter generating, state=%d", gv.state)
	}
	staleEpoch := gv.workEpoch

	// ctrl+c stays the hard-quit even mid-generation.
	if _, cmd := gv.Update(tea.KeyMsg{Type: tea.KeyCtrlC}); cmd == nil {
		t.Error("ctrl+c during generation should quit the app")
	}

	stopped, cmd := gv.Update(tea.KeyMsg{Type: tea.KeyEsc})
	sv := stopped.(Model)
	if cmd != nil {
		t.Error("esc during generation should not quit the app")
	}
	if sv.state != stateHome {
		t.Errorf("cancel should return to home, state=%d", sv.state)
	}
	if sv.workEpoch == staleEpoch {
		t.Error("cancel should bump the work epoch")
	}

	// The in-flight result now arrives late; it must be ignored.
	late, _ := sv.Update(genDoneMsg{kind: "guide", md: "# Study Guide", epoch: staleEpoch})
	lv := late.(Model)
	if lv.state != stateHome {
		t.Errorf("stale result should be ignored, state=%d", lv.state)
	}
	if _, err := os.Stat(ws.GuidePath("algebra")); err == nil {
		t.Error("cancelled generation must not write an artifact")
	}
}

func TestFitCell(t *testing.T) {
	if got := fitCell("ab", 5); got != "ab   " {
		t.Errorf("short should pad to width: %q", got)
	}
	long := fitCell("merise-dr-ouedraogo-complet", listNameCol)
	if n := len([]rune(long)); n != listNameCol {
		t.Errorf("long should clamp to %d cells, got %d (%q)", listNameCol, n, long)
	}
	if !strings.HasSuffix(long, "…") {
		t.Errorf("long should be ellipsized: %q", long)
	}
}

func TestCourseDetailPanel(t *testing.T) {
	courses := []workspace.Course{{Name: "algebra", HasGuide: true}}
	m := New("claude", nil, workspace.Workspace{Root: "/s"}, courses)
	mm, _ := m.Update(tea.WindowSizeMsg{Width: 100, Height: 40})
	out := mm.View()

	if !strings.Contains(out, "ALGEBRA") {
		t.Errorf("detail panel missing course title")
	}
	if !strings.Contains(out, "ready") {
		t.Errorf("guide-ready marker missing for course with a guide")
	}
	if !strings.Contains(out, "build q&a") {
		t.Errorf("missing-qa build hint absent: %q", out)
	}
}

func TestHelpOverlay(t *testing.T) {
	courses := []workspace.Course{{Name: "algebra"}}
	m := New("claude", nil, workspace.Workspace{Root: "/x"}, courses)
	mm, _ := m.Update(tea.WindowSizeMsg{Width: 100, Height: 30})

	h, _ := mm.(Model).Update(keyRunes("?"))
	if h.(Model).state != stateHelp {
		t.Fatalf("? should open help, state=%d", h.(Model).state)
	}
	if !strings.Contains(h.View(), "revise") {
		t.Errorf("help overlay missing key reference: %q", h.View())
	}

	back, _ := h.(Model).Update(keyRunes("x"))
	if back.(Model).state != stateHome {
		t.Errorf("any key should close help, state=%d", back.(Model).state)
	}
}

func TestWindow(t *testing.T) {
	cases := []struct{ n, cur, max, ws, we int }{
		{10, 0, 5, 0, 5},
		{10, 9, 5, 5, 10},
		{3, 1, 5, 0, 3},
	}
	for _, c := range cases {
		if s, e := window(c.n, c.cur, c.max); s != c.ws || e != c.we {
			t.Errorf("window(%d,%d,%d)=(%d,%d) want (%d,%d)", c.n, c.cur, c.max, s, e, c.ws, c.we)
		}
	}
}

func TestIngestBrowser(t *testing.T) {
	dir := t.TempDir()
	mustMkdir(t, filepath.Join(dir, "sub"))
	for _, f := range []string{"a.pdf", "b.txt", "c.png", ".hidden.pdf"} {
		mustWrite(t, filepath.Join(dir, f), "x")
	}

	m := New("claude", nil, workspace.Workspace{}, nil)
	m.ingSelected = map[string]bool{}
	m.loadDir(dir)

	// sub/ (dir, first) + a.pdf + b.txt; .png and dotfile excluded.
	if len(m.ingEntries) != 3 {
		t.Fatalf("want 3 entries, got %d: %+v", len(m.ingEntries), m.ingEntries)
	}
	if !m.ingEntries[0].isDir || m.ingEntries[0].name != "sub" {
		t.Errorf("directories should sort first, got %+v", m.ingEntries[0])
	}

	// Select a.pdf (index 1) with space.
	m.ingCursor = 1
	sel, _ := m.updateIngestPick(keyRunes(" "))
	if !sel.(Model).ingSelected[filepath.Join(dir, "a.pdf")] {
		t.Errorf("space should select the current file")
	}

	// Enter on the directory descends into it.
	m.ingCursor = 0
	desc, _ := m.updateIngestPick(tea.KeyMsg{Type: tea.KeyEnter})
	if desc.(Model).ingDir != filepath.Join(dir, "sub") {
		t.Errorf("enter on a dir should descend, ingDir=%q", desc.(Model).ingDir)
	}
}

func TestIngestOptsValidation(t *testing.T) {
	m := New("claude", nil, workspace.Workspace{}, nil)
	m.ingSelected = map[string]bool{"/x/a.pdf": true, "/x/b.pdf": true}
	m.ingKind = "course"
	m.state = stateIngestOpts
	m.ingInput.SetValue("")

	res, _ := m.updateIngestOpts(tea.KeyMsg{Type: tea.KeyEnter})
	rm := res.(Model)
	if rm.state != stateIngestOpts {
		t.Errorf("multi-file ingest with no name should stay on the form, state=%d", rm.state)
	}
	if !strings.Contains(rm.notice, "course name") {
		t.Errorf("want a name-required notice, got %q", rm.notice)
	}
}

func TestIngestDoneRefresh(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("GENIUS_HOME", dir)
	ws, err := workspace.Open(workspace.Config{})
	if err != nil {
		t.Fatal(err)
	}
	mustMkdir(t, ws.Path("courses", "algebra"))

	m := New("claude", nil, ws, nil)
	res, _ := m.ingestDone(ingestDoneMsg{course: "algebra", ingested: 2})
	rm := res.(Model)
	if rm.state != stateHome {
		t.Errorf("ingest should land home, state=%d", rm.state)
	}
	if len(rm.courses) != 1 || rm.courses[0].Name != "algebra" {
		t.Errorf("course list should refresh to show algebra, got %+v", rm.courses)
	}
	if !strings.Contains(rm.notice, "ingested 2") {
		t.Errorf("want an ingest summary, got %q", rm.notice)
	}
}

// A partial batch must report the files that landed AND every failure — an error
// on one document used to mask the successes entirely.
func TestIngestDonePartialFailure(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("GENIUS_HOME", dir)
	ws, err := workspace.Open(workspace.Config{})
	if err != nil {
		t.Fatal(err)
	}
	mustMkdir(t, ws.Path("courses", "algebra"))

	m := New("claude", nil, ws, nil)
	res, _ := m.ingestDone(ingestDoneMsg{
		course:   "algebra",
		ingested: 2,
		errs:     []string{"chap3.pdf: boom", "chap4.pdf: kaput"},
	})
	rm := res.(Model)

	if !strings.Contains(rm.notice, "ingested 2") {
		t.Errorf("partial batch must still report successes, got %q", rm.notice)
	}
	if !strings.Contains(rm.notice, "2 failed") {
		t.Errorf("partial batch must report the failure count, got %q", rm.notice)
	}
	if !strings.Contains(rm.notice, "chap3.pdf") {
		t.Errorf("partial batch must name a failed file, got %q", rm.notice)
	}
	if rm.noticeLvl != lvlWarn {
		t.Errorf("partial failure should warn (lvlWarn=%d), got %d", lvlWarn, rm.noticeLvl)
	}
}

func TestChaptersScopedGenerate(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("GENIUS_HOME", dir)
	ws, err := workspace.Open(workspace.Config{})
	if err != nil {
		t.Fatal(err)
	}
	mustMkdir(t, ws.Path("courses", "algebra"))
	mustWrite(t, ws.Path("courses", "algebra", "chap01.md"), "# A")
	mustWrite(t, ws.Path("courses", "algebra", "chap02.md"), "# B")

	m := New("claude", nil, ws, []workspace.Course{{Name: "algebra"}})
	mm, _ := m.Update(tea.WindowSizeMsg{Width: 100, Height: 30})

	c, _ := mm.(Model).Update(keyRunes("f"))
	cm := c.(Model)
	if cm.state != stateChapters {
		t.Fatalf("f should open the chapter picker, state=%d", cm.state)
	}
	if len(cm.chapFiles) != 2 {
		t.Fatalf("want 2 chapters, got %d", len(cm.chapFiles))
	}

	sel, _ := cm.Update(keyRunes(" "))
	sm := sel.(Model)
	if got := sm.scopedChapters(); len(got) != 1 || got[0] != "chap01.md" {
		t.Errorf("scopedChapters should reflect the toggle, got %v", got)
	}

	// g with a nil engine reports the guard (proves the key routes to generate).
	g, _ := sm.Update(keyRunes("g"))
	if !strings.Contains(g.(Model).notice, "no engine") {
		t.Errorf("g should route to scoped generate, notice=%q", g.(Model).notice)
	}
}

func TestResolveGenTarget(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("GENIUS_HOME", dir)
	ws, err := workspace.Open(workspace.Config{})
	if err != nil {
		t.Fatal(err)
	}
	mustMkdir(t, ws.Path("courses", "algebra"))
	for _, f := range []string{"chap01.md", "chap02.md", "chap03.md"} {
		mustWrite(t, ws.Path("courses", "algebra", f), "# x")
	}
	m := New("claude", nil, ws, nil)

	// no selection → whole course
	if p, _ := m.resolveGenTarget("guide", "algebra", nil); p != ws.GuidePath("algebra") {
		t.Errorf("nil files should target the whole guide, got %s", p)
	}
	// every chapter selected → whole course
	all := []string{"chap01.md", "chap02.md", "chap03.md"}
	if p, _ := m.resolveGenTarget("qa", "algebra", all); p != ws.QAPath("algebra") {
		t.Errorf("all-selected should target the whole qa, got %s", p)
	}
	// one chapter → scoped artifact
	if p, _ := m.resolveGenTarget("guide", "algebra", []string{"chap01.md"}); p != ws.ChapterGuidePath("algebra", "chap01") {
		t.Errorf("single chapter should target a scoped guide, got %s", p)
	}
	// a span → combined scoped artifact
	if p, _ := m.resolveGenTarget("guide", "algebra", []string{"chap01.md", "chap02.md"}); p != ws.ChapterGuidePath("algebra", "chap01+chap02") {
		t.Errorf("span should target a combined guide, got %s", p)
	}
}

func TestGenDoneWritesChapterPath(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("GENIUS_HOME", dir)
	ws, err := workspace.Open(workspace.Config{})
	if err != nil {
		t.Fatal(err)
	}
	mustMkdir(t, ws.Path("courses", "algebra"))

	m := New("claude", nil, ws, []workspace.Course{{Name: "algebra"}})
	mm, _ := m.Update(tea.WindowSizeMsg{Width: 100, Height: 30})
	gm := mm.(Model)
	gm.genCourse = "algebra"
	gm.genKind = "guide"
	gm.genPath = ws.ChapterGuidePath("algebra", "chap01")
	gm.genTitle = "algebra/chap01 · guide"

	res, _ := gm.genDone(genDoneMsg{kind: "guide", md: "# scoped guide"})
	rm := res.(Model)

	if _, err := os.Stat(ws.ChapterGuidePath("algebra", "chap01")); err != nil {
		t.Errorf("scoped guide not written: %v", err)
	}
	if _, err := os.Stat(ws.GuidePath("algebra")); err == nil {
		t.Errorf("whole-course guide must be left untouched")
	}
	if rm.state != stateReader {
		t.Errorf("genDone should open the reader, state=%d", rm.state)
	}
	if rm.courses[0].GuideCount() != 1 {
		t.Errorf("chip count should reflect the new scoped guide, got %d", rm.courses[0].GuideCount())
	}
}

func TestQuizSourcePicker(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("GENIUS_HOME", dir)
	ws, err := workspace.Open(workspace.Config{})
	if err != nil {
		t.Fatal(err)
	}
	mustMkdir(t, ws.Path("courses", "algebra"))
	const qa = "## Q1. What is x?\n\nThe answer.\n"

	// Whole + one scoped → picker with whole / chapter / merged.
	mustMkdir(t, ws.Path("qa", "algebra"))
	mustWrite(t, ws.Path("qa", "algebra.md"), qa)
	mustWrite(t, ws.Path("qa", "algebra", "chap01.md"), qa)

	m := New("claude", nil, ws, []workspace.Course{{Name: "algebra"}})
	res, _ := m.startQuiz()
	rm := res.(Model)
	if rm.state != stateQuizPick {
		t.Fatalf("multiple Q&A should open the picker, state=%d", rm.state)
	}
	if len(rm.quizPick) != 3 {
		t.Fatalf("want whole + chapter + merged = 3 sources, got %d (%+v)", len(rm.quizPick), rm.quizPick)
	}
	if !rm.quizPick[len(rm.quizPick)-1].merged {
		t.Errorf("last source should be the merged entry")
	}

	// Selecting a source enters the quiz.
	q, _ := rm.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if q.(Model).state != stateQuiz {
		t.Errorf("picking a source should enter the quiz, state=%d", q.(Model).state)
	}
}

func TestQuizSingleSourceSkipsPicker(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("GENIUS_HOME", dir)
	ws, err := workspace.Open(workspace.Config{})
	if err != nil {
		t.Fatal(err)
	}
	mustMkdir(t, ws.Path("courses", "algebra"))
	// Only the whole-course Q&A exists → load straight into the quiz.
	mustWrite(t, ws.Path("qa", "algebra.md"), "## Q1. What is x?\n\nThe answer.\n")

	m := New("claude", nil, ws, []workspace.Course{{Name: "algebra"}})
	res, _ := m.startQuiz()
	if res.(Model).state != stateQuiz {
		t.Errorf("a single Q&A source should skip the picker, state=%d", res.(Model).state)
	}
}

// The chapter hub and quiz picker must render without panicking and stay within
// the one-row-status-bar layout.
func TestNewScreensRender(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("GENIUS_HOME", dir)
	ws, err := workspace.Open(workspace.Config{})
	if err != nil {
		t.Fatal(err)
	}
	mustMkdir(t, ws.Path("courses", "algebra"))
	mustWrite(t, ws.Path("courses", "algebra", "chap01.md"), "# A")
	mustWrite(t, ws.Path("courses", "algebra", "chap02.md"), "# B")
	mustMkdir(t, ws.Path("guides", "algebra"))
	mustWrite(t, ws.Path("guides", "algebra", "chap01.md"), "# scoped")

	m := New("claude", nil, ws, []workspace.Course{{Name: "algebra"}})
	base, _ := m.Update(tea.WindowSizeMsg{Width: 100, Height: 30})

	hub, _ := base.(Model).Update(keyRunes("f"))
	if got := hub.(Model).View(); !strings.Contains(got, "chapters") {
		t.Errorf("chapter hub should render its title, got %q", firstLine(got))
	}

	pick := base.(Model)
	pick.state = stateQuizPick
	pick.quizName = "algebra"
	pick.quizPick = []quizSource{{label: "whole course"}, {label: "chap01"}, {label: "all chapters merged", merged: true}}
	if got := pick.View(); !strings.Contains(got, "revise") {
		t.Errorf("quiz picker should render its title, got %q", firstLine(got))
	}
}

func firstLine(s string) string {
	if i := strings.IndexByte(s, '\n'); i >= 0 {
		return s[:i]
	}
	return s
}

func mustMkdir(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(path, 0o755); err != nil {
		t.Fatal(err)
	}
}

func mustWrite(t *testing.T, path, body string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

// Merged quiz sources restart their Q numbering per file; the progress line
// must show the session position, not the pair's own N — and a large N from a
// malformed file (## Q1000.) must not crash the renderer.
func TestQuizProgressUsesSessionPosition(t *testing.T) {
	m := New("claude", nil, workspace.Workspace{Root: "/x"}, nil)
	mm, _ := m.Update(tea.WindowSizeMsg{Width: 100, Height: 30})
	qm := mm.(Model).enterQuiz("algebra", []quiz.Pair{
		{N: 1, Question: "a?", Answer: "a"},
		{N: 1000, Question: "b?", Answer: "b"}, // repeated/large N from a merged file
	})

	if out := qm.viewQuiz(); !strings.Contains(out, "Q 1/2") {
		t.Errorf("first question should show Q 1/2, got %q", out)
	}
	qm = qm.nextQuestion()
	if out := qm.viewQuiz(); !strings.Contains(out, "Q 2/2") {
		t.Errorf("second question should show Q 2/2 (not the pair's N), got %q", out)
	}
}

func TestGradientColorAtBounds(t *testing.T) {
	if got := gradientColorAt(0); string(got) != "#6BA8F5" {
		t.Errorf("t=0 want #6BA8F5 got %s", got)
	}
	if got := gradientColorAt(1); string(got) != "#E66FB0" {
		t.Errorf("t=1 want #E66FB0 got %s", got)
	}
}

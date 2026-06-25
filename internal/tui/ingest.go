package tui

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/mibienpanjoe/genius/internal/convert"
	"github.com/mibienpanjoe/genius/internal/engine"
	"github.com/mibienpanjoe/genius/internal/workspace"
)

// dirEntry is one row in the ingest file browser: a sub-directory to descend
// into or an ingestable document to pick.
type dirEntry struct {
	name  string
	path  string
	isDir bool
}

// ingestExts are the document types markitdown can convert; the browser hides
// everything else so the list stays relevant.
var ingestExts = map[string]bool{
	".pdf": true, ".pptx": true, ".ppt": true, ".docx": true, ".doc": true,
	".md": true, ".txt": true, ".html": true, ".htm": true, ".csv": true,
	".xlsx": true, ".epub": true,
}

func ingestable(name string) bool {
	return ingestExts[strings.ToLower(filepath.Ext(name))]
}

// ingestJob is a resolved batch handed to the converter goroutine.
type ingestJob struct {
	files     []string
	kind      string // "course" | "exercise"
	target    string // course name (course kind) or target course (exercise kind)
	overwrite bool
	opts      convert.IngestOpts
}

// ingestDoneMsg reports a finished batch back to the UI thread.
type ingestDoneMsg struct {
	course   string
	ingested int
	skipped  int
	firstErr string
}

// openIngest enters the file browser, refusing up front when the converter is
// absent (same guard as the CLI, ERR-031).
func (m Model) openIngest() (tea.Model, tea.Cmd) {
	if !convert.Available() {
		m.notice = convert.InstallHint
		m.noticeLvl = lvlErr
		return m, nil
	}
	start := m.ingDir
	if start == "" {
		if home, err := os.UserHomeDir(); err == nil {
			start = home
		} else {
			start, _ = os.Getwd()
		}
	}
	m.ingSelected = map[string]bool{}
	m.loadDir(start)
	m.state = stateIngestPick
	return m, nil
}

// loadDir reads dir into the browser: sub-directories first, then ingestable
// files, both sorted; dotfiles are hidden.
func (m *Model) loadDir(dir string) {
	m.ingDir = dir
	m.ingCursor = 0
	m.ingEntries = nil

	entries, err := os.ReadDir(dir)
	if err != nil {
		return
	}
	var dirs, files []dirEntry
	for _, e := range entries {
		name := e.Name()
		if strings.HasPrefix(name, ".") {
			continue
		}
		full := filepath.Join(dir, name)
		if e.IsDir() {
			dirs = append(dirs, dirEntry{name: name, path: full, isDir: true})
		} else if ingestable(name) {
			files = append(files, dirEntry{name: name, path: full})
		}
	}
	sort.Slice(dirs, func(i, j int) bool { return dirs[i].name < dirs[j].name })
	sort.Slice(files, func(i, j int) bool { return files[i].name < files[j].name })
	m.ingEntries = append(dirs, files...)
}

func (m Model) updateIngestPick(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c":
		return m, tea.Quit
	case "esc", "q":
		m.state = stateHome
		return m, nil
	case "up", "k":
		if m.ingCursor > 0 {
			m.ingCursor--
		}
	case "down", "j":
		if m.ingCursor < len(m.ingEntries)-1 {
			m.ingCursor++
		}
	case "backspace", "left", "h":
		if parent := filepath.Dir(m.ingDir); parent != m.ingDir {
			m.loadDir(parent)
		}
	case " ":
		if e := m.currentEntry(); e != nil && !e.isDir {
			m.ingSelected[e.path] = !m.ingSelected[e.path]
		}
	case "enter", "right", "l":
		e := m.currentEntry()
		if e == nil {
			return m, nil
		}
		if e.isDir {
			m.loadDir(e.path)
			return m, nil
		}
		if len(m.selectedFiles()) == 0 {
			m.ingSelected[e.path] = true
		}
		return m.openIngestOpts()
	}
	return m, nil
}

func (m Model) currentEntry() *dirEntry {
	if m.ingCursor < 0 || m.ingCursor >= len(m.ingEntries) {
		return nil
	}
	return &m.ingEntries[m.ingCursor]
}

func (m Model) selectedFiles() []string {
	var fs []string
	for p, on := range m.ingSelected {
		if on {
			fs = append(fs, p)
		}
	}
	sort.Strings(fs)
	return fs
}

// openIngestOpts moves to the options form, prefilling the course name from a
// single file's slug.
func (m Model) openIngestOpts() (tea.Model, tea.Cmd) {
	sel := m.selectedFiles()
	m.ingInput.SetValue("")
	if len(sel) == 1 && m.ingKind == "course" {
		m.ingInput.SetValue(workspace.Slug(sel[0]))
	}
	m.ingField = 0
	m.ingInput.Focus()
	m.state = stateIngestOpts
	return m, nil
}

func (m Model) updateIngestOpts(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c":
		return m, tea.Quit
	case "esc":
		m.ingInput.Blur()
		m.notice = ""
		m.state = stateIngestPick
		return m, nil
	case "tab":
		if m.ingKind == "course" {
			m.ingKind = "exercise"
			m.ingInput.Placeholder = "target course"
		} else {
			m.ingKind = "course"
			m.ingInput.Placeholder = "course name"
		}
		return m, nil
	case "up":
		if m.ingField > 0 {
			m.ingField--
		}
		m.syncInputFocus()
		return m, nil
	case "down":
		if m.ingField < 2 {
			m.ingField++
		}
		m.syncInputFocus()
		return m, nil
	case "enter":
		return m.runIngest()
	}

	switch m.ingField {
	case 0:
		var cmd tea.Cmd
		m.ingInput, cmd = m.ingInput.Update(msg)
		return m, cmd
	case 1:
		if k := msg.String(); k == " " || k == "left" || k == "right" {
			m.ingDescribe = !m.ingDescribe
		}
	case 2:
		if k := msg.String(); k == " " || k == "left" || k == "right" {
			m.ingOverlay = !m.ingOverlay
		}
	}
	return m, nil
}

func (m *Model) syncInputFocus() {
	if m.ingField == 0 {
		m.ingInput.Focus()
	} else {
		m.ingInput.Blur()
	}
}

// runIngest validates the form and launches the converter goroutine.
func (m Model) runIngest() (tea.Model, tea.Cmd) {
	sel := m.selectedFiles()
	if len(sel) == 0 {
		m.notice = "no files selected"
		m.noticeLvl = lvlWarn
		m.state = stateHome
		return m, nil
	}
	target := strings.TrimSpace(m.ingInput.Value())
	switch m.ingKind {
	case "course":
		if target == "" {
			if len(sel) == 1 {
				target = workspace.Slug(sel[0])
			} else {
				m.notice = "type a course name for a multi-file ingest"
				m.noticeLvl = lvlWarn
				return m, nil
			}
		}
	case "exercise":
		if target == "" {
			m.notice = "an exercise set needs a target course"
			m.noticeLvl = lvlWarn
			return m, nil
		}
	}

	m.ingInput.Blur()
	m.notice = ""
	m.state = stateIngesting
	job := ingestJob{
		files:     sel,
		kind:      m.ingKind,
		target:    target,
		overwrite: m.ingOverlay,
		opts:      convert.IngestOpts{Describe: m.ingDescribe, MinImagePx: convert.DefaultMinImagePx},
	}
	if m.reduceMotion {
		return m, ingestCmd(m.eng, m.ws, job)
	}
	return m, tea.Batch(m.spinner.Tick, ingestCmd(m.eng, m.ws, job))
}

// ingestCmd converts and files each selected document off the UI goroutine.
func ingestCmd(eng engine.Engine, ws workspace.Workspace, job ingestJob) tea.Cmd {
	return func() tea.Msg {
		var ingested, skipped int
		var firstErr string
		course := job.target

		for _, f := range job.files {
			var target, assets string
			switch job.kind {
			case "course":
				name := job.target
				if name == "" {
					name = workspace.Slug(f)
				}
				course = name
				target = ws.CourseDocPath(name, workspace.Slug(f))
				assets = ws.Path("courses", name, "assets")
			case "exercise":
				target = ws.ExerciseSetPath(job.target, workspace.Slug(f))
				assets = ws.Path("exercises", job.target, "assets")
			}

			res, err := convert.Ingest(context.Background(), f, assets, eng, job.opts)
			if err != nil {
				if firstErr == "" {
					firstErr = err.Error()
				}
				continue
			}
			werr := ws.WriteArtifact(target, []byte(res.Markdown+"\n"), job.overwrite)
			switch {
			case errors.Is(werr, workspace.ErrExists):
				skipped++
			case werr != nil:
				if firstErr == "" {
					firstErr = werr.Error()
				}
			default:
				ingested++
			}
		}
		return ingestDoneMsg{course: course, ingested: ingested, skipped: skipped, firstErr: firstErr}
	}
}

// ingestDone refreshes the course list, points the cursor at the affected
// course, and reports a summary.
func (m Model) ingestDone(msg ingestDoneMsg) (tea.Model, tea.Cmd) {
	if cs, err := m.ws.Courses(); err == nil {
		m.courses = cs
		for i, c := range m.courses {
			if c.Name == msg.course {
				m.cursor = i
				break
			}
		}
		if m.cursor >= len(m.courses) {
			m.cursor = 0
		}
	}

	switch {
	case msg.firstErr != "":
		m.notice = "ingest: " + msg.firstErr
		m.noticeLvl = lvlErr
	case msg.ingested == 0 && msg.skipped > 0:
		m.notice = fmt.Sprintf("nothing ingested — %d already existed (toggle overwrite to replace)", msg.skipped)
		m.noticeLvl = lvlWarn
	default:
		summary := fmt.Sprintf("ingested %d file(s) into %s", msg.ingested, msg.course)
		if msg.skipped > 0 {
			summary += fmt.Sprintf(" · %d skipped (exists)", msg.skipped)
		}
		m.notice = summary
		m.noticeLvl = lvlInfo
	}
	m.state = stateHome
	return m, nil
}

func (m Model) viewIngestPick() string {
	var b strings.Builder
	b.WriteString(styleTitle.Render("ingest") + styleMuted.Render("  "+m.ingDir))
	b.WriteString("\n\n")

	if len(m.ingEntries) == 0 {
		b.WriteString(styleMuted.Render("(no folders or ingestable documents here — ⌫ to go up)"))
		b.WriteString("\n\n" + styleMuted.Render("⌫ up · esc back"))
		return lipgloss.NewStyle().Padding(1, 0, 0, 3).Render(b.String())
	}

	start, end := window(len(m.ingEntries), m.ingCursor, m.contentHeight()-6)
	if start > 0 {
		b.WriteString(styleMuted.Render("   ↑ more") + "\n")
	}
	for i := start; i < end; i++ {
		e := m.ingEntries[i]
		cursor := "  "
		if i == m.ingCursor {
			cursor = lipgloss.NewStyle().Foreground(cPrimary).Render("▌ ")
		}
		box := "   "
		label := e.name
		if e.isDir {
			label += "/"
		} else {
			box = "[ ]"
			if m.ingSelected[e.path] {
				box = lipgloss.NewStyle().Foreground(cSuccess).Render("[x]")
			}
		}
		nameStyle := styleBody
		if e.isDir {
			nameStyle = styleInfo
		}
		if i == m.ingCursor {
			nameStyle = lipgloss.NewStyle().Bold(true).Foreground(cText)
		}
		b.WriteString(cursor + box + " " + nameStyle.Render(label) + "\n")
	}
	if end < len(m.ingEntries) {
		b.WriteString(styleMuted.Render("   ↓ more") + "\n")
	}

	n := len(m.selectedFiles())
	footer := "enter open/pick · space select · ⌫ up dir · esc back"
	if n > 0 {
		footer = fmt.Sprintf("%d selected · ", n) + footer
	}
	b.WriteString("\n" + styleMuted.Render(footer))
	return lipgloss.NewStyle().Padding(1, 0, 0, 3).Render(b.String())
}

func (m Model) viewIngestOpts() string {
	sel := m.selectedFiles()
	names := make([]string, len(sel))
	for i, p := range sel {
		names[i] = filepath.Base(p)
	}

	var b strings.Builder
	b.WriteString(styleTitle.Render("ingest options"))
	b.WriteString("\n\n")
	b.WriteString(styleMuted.Render(fmt.Sprintf("%d file(s): ", len(sel))) +
		styleBody.Render(fitCell(strings.Join(names, ", "), 60)))
	b.WriteString("\n\n")

	b.WriteString(styleMuted.Render("kind  ") + kindToggle(m.ingKind) +
		styleMuted.Render("   (tab to switch)"))
	b.WriteString("\n\n")

	nameLabel := "course name"
	if m.ingKind == "exercise" {
		nameLabel = "target course"
	}
	b.WriteString(fieldRow(m.ingField == 0, nameLabel, m.ingInput.View()))
	b.WriteString("\n")
	b.WriteString(fieldRow(m.ingField == 1, "describe images", onOff(m.ingDescribe)))
	b.WriteString("\n")
	b.WriteString(fieldRow(m.ingField == 2, "overwrite existing", onOff(m.ingOverlay)))
	b.WriteString("\n\n")
	b.WriteString(styleKey.Render("enter") + styleMuted.Render(" ingest · ") +
		styleKey.Render("esc") + styleMuted.Render(" back"))

	if m.notice != "" {
		b.WriteString("\n\n" + noticeLine(m.noticeLvl, m.notice))
	}
	return lipgloss.NewStyle().Padding(1, 0, 0, 3).Render(b.String())
}

func (m Model) viewIngesting() string {
	body := m.spinnerHead() + "ingesting " + itoa(len(m.selectedFiles())) +
		" file(s) with markitdown…\n\n" +
		styleMuted.Render("converting to markdown and extracting figures")
	return lipgloss.NewStyle().Padding(2, 0, 0, 3).Render(body)
}

// fieldRow renders one options-form row, marking the active field.
func fieldRow(active bool, label, value string) string {
	bar := "  "
	labelStyle := styleMuted
	if active {
		bar = lipgloss.NewStyle().Foreground(cPrimary).Render("▌ ")
		labelStyle = lipgloss.NewStyle().Bold(true).Foreground(cText)
	}
	return bar + labelStyle.Render(fitCell(label, 20)) + value
}

func kindToggle(kind string) string {
	course, exercise := styleMuted.Render("course"), styleMuted.Render("exercise")
	on := lipgloss.NewStyle().Bold(true).Foreground(cPrimary)
	if kind == "exercise" {
		exercise = on.Render("exercise")
	} else {
		course = on.Render("course")
	}
	return course + styleMuted.Render(" / ") + exercise
}

func onOff(b bool) string {
	if b {
		return lipgloss.NewStyle().Foreground(cSuccess).Render("on")
	}
	return styleMuted.Render("off")
}

// window returns the [start,end) slice of an n-item list to show around cursor
// within at most max rows.
func window(n, cursor, max int) (int, int) {
	if max < 3 {
		max = 3
	}
	if n <= max {
		return 0, n
	}
	start := cursor - max/2
	if start < 0 {
		start = 0
	}
	end := start + max
	if end > n {
		end = n
		start = end - max
	}
	return start, end
}

package workspace

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// Workspace is the single rooted study home (INV-01). All study artifacts live
// under Root in the courses/ guides/ qa/ exercises/ tree.
type Workspace struct {
	Root string
}

// subdirs are the top-level workspace directories created on first run (FR-012).
var subdirs = []string{"courses", "guides", "qa", "exercises"}

// ResolveRoot determines the workspace root by precedence (FR-011):
// $GENIUS_HOME → config study_root → ~/study.
func ResolveRoot(cfg Config) string {
	if v := os.Getenv("GENIUS_HOME"); v != "" {
		return v
	}
	if cfg.StudyRoot != "" {
		return expandHome(cfg.StudyRoot)
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "study"
	}
	return filepath.Join(home, "study")
}

// Open resolves the root, creates the workspace tree if missing (FR-012,
// usability: first run never blocks), and returns the Workspace.
func Open(cfg Config) (Workspace, error) {
	root := ResolveRoot(cfg)
	w := Workspace{Root: root}
	for _, d := range subdirs {
		if err := os.MkdirAll(filepath.Join(root, d), 0o755); err != nil {
			return w, err
		}
	}
	return w, nil
}

// Path joins workspace-relative parts onto the root.
func (w Workspace) Path(parts ...string) string {
	return filepath.Join(append([]string{w.Root}, parts...)...)
}

// Course is one study subject: a directory under courses/ (INV-02), with
// derived artifact counts for the home dashboard (FR-022).
type Course struct {
	Name         string
	HasGuide     bool // guides/<name>.md exists
	HasQA        bool // qa/<name>.md exists
	ExerciseSets int  // count of *.md under exercises/<name>/
}

// GuideCount / QACount expose the booleans as chip counts (g·N q·N e·N).
func (c Course) GuideCount() int {
	if c.HasGuide {
		return 1
	}
	return 0
}

func (c Course) QACount() int {
	if c.HasQA {
		return 1
	}
	return 0
}

// Courses scans courses/ and returns each course with its artifact counts,
// sorted by name. Each directory under courses/ is one course (FR-013).
func (w Workspace) Courses() ([]Course, error) {
	entries, err := os.ReadDir(w.Path("courses"))
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var courses []Course
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		name := e.Name()
		courses = append(courses, Course{
			Name:         name,
			HasGuide:     fileExists(w.Path("guides", name+".md")),
			HasQA:        fileExists(w.Path("qa", name+".md")),
			ExerciseSets: countMarkdown(w.Path("exercises", name)),
		})
	}
	sort.Slice(courses, func(i, j int) bool {
		return courses[i].Name < courses[j].Name
	})
	return courses, nil
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

// countMarkdown counts *.md files directly under dir (exercise sets).
func countMarkdown(dir string) int {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return 0
	}
	n := 0
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".md") {
			n++
		}
	}
	return n
}

// expandHome expands a leading ~ to the user home directory.
func expandHome(p string) string {
	if p == "~" || strings.HasPrefix(p, "~/") {
		if home, err := os.UserHomeDir(); err == nil {
			return filepath.Join(home, strings.TrimPrefix(p, "~"))
		}
	}
	return p
}

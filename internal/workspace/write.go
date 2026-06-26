package workspace

import (
	"errors"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

// ErrExists is returned when a write target already exists and force is false,
// so callers can apply the overwrite-confirmation rule (INV-03, FR-034).
var ErrExists = errors.New("output file already exists")

// ErrNoMaterial is returned when a course has no source markdown to ground on
// (INV-05/ERR-041); generation must refuse rather than fabricate.
var ErrNoMaterial = errors.New("course has no source material")

// CourseDocPath is where an ingested course document lands: courses/<name>/<base>.md.
func (w Workspace) CourseDocPath(name, base string) string {
	return w.Path("courses", name, base+".md")
}

// ExerciseSetPath is where an ingested exercise set lands:
// exercises/<course>/<set>.md.
func (w Workspace) ExerciseSetPath(course, set string) string {
	return w.Path("exercises", course, set+".md")
}

// GuidePath / QAPath are the whole-course artifact locations (INV-02).
func (w Workspace) GuidePath(name string) string { return w.Path("guides", name+".md") }
func (w Workspace) QAPath(name string) string    { return w.Path("qa", name+".md") }

// ChapterGuidePath / ChapterQAPath are the scoped artifact locations: a guide or
// Q&A grounded on one chapter (or a span), filed under the course's own subdir so
// it never overwrites the whole-course slot. scope is the joined chapter slug
// (see scopeName in the TUI), e.g. "chap01" or "chap01+chap02".
func (w Workspace) ChapterGuidePath(course, scope string) string {
	return w.Path("guides", course, scope+".md")
}
func (w Workspace) ChapterQAPath(course, scope string) string {
	return w.Path("qa", course, scope+".md")
}

// ScopeName builds the scoped-artifact basename from chapter filenames: each
// file's slug (sans .md), sorted, joined with "+". chap02.md,chap01.md ->
// "chap01+chap02".
func ScopeName(files []string) string {
	parts := make([]string, len(files))
	for i, f := range files {
		parts[i] = Slug(f)
	}
	sort.Strings(parts)
	return strings.Join(parts, "+")
}

// scopeIsWhole reports whether a chapter selection means the whole course: an
// empty selection, or every chapter selected.
func (w Workspace) scopeIsWhole(course string, files []string) bool {
	if len(files) == 0 {
		return true
	}
	all, err := w.CourseFiles(course)
	return err == nil && len(all) > 0 && len(files) == len(all)
}

// GuideTarget / QATarget resolve where a generation should be written: the
// whole-course slot for an empty/all selection, else a scoped artifact under the
// course subdir. Shared by the CLI and the TUI so both agree on placement.
func (w Workspace) GuideTarget(course string, files []string) string {
	if w.scopeIsWhole(course, files) {
		return w.GuidePath(course)
	}
	return w.ChapterGuidePath(course, ScopeName(files))
}

func (w Workspace) QATarget(course string, files []string) string {
	if w.scopeIsWhole(course, files) {
		return w.QAPath(course)
	}
	return w.ChapterQAPath(course, ScopeName(files))
}

// WriteArtifact writes data to path, creating parent dirs. It refuses to
// overwrite an existing file unless force is true (INV-03); on refusal it
// returns ErrExists and leaves the existing file untouched.
func (w Workspace) WriteArtifact(path string, data []byte, force bool) error {
	if !force {
		if _, err := os.Stat(path); err == nil {
			return ErrExists
		}
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

var slugInvalid = regexp.MustCompile(`[^a-z0-9]+`)

// Slug derives a stable course/set identifier from a filename or title:
// lowercased, non-alphanumerics collapsed to single hyphens, trimmed.
func Slug(s string) string {
	base := filepath.Base(s)
	if ext := filepath.Ext(base); ext != "" {
		base = strings.TrimSuffix(base, ext)
	}
	base = strings.ToLower(base)
	base = slugInvalid.ReplaceAllString(base, "-")
	base = strings.Trim(base, "-")
	if base == "" {
		return "untitled"
	}
	return base
}

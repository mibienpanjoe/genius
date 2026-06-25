package workspace

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// CourseMaterial concatenates all markdown under courses/<name>/ (sorted by
// filename) into one grounding blob, each file prefixed with its name. It
// returns ErrNoMaterial when the course has no source markdown (INV-05/ERR-041).
func (w Workspace) CourseMaterial(name string) (string, error) {
	dir := w.Path("courses", name)
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return "", ErrNoMaterial
		}
		return "", err
	}
	var files []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".md") {
			files = append(files, e.Name())
		}
	}
	sort.Strings(files)
	if len(files) == 0 {
		return "", ErrNoMaterial
	}

	var b strings.Builder
	for _, f := range files {
		data, err := os.ReadFile(filepath.Join(dir, f))
		if err != nil {
			return "", err
		}
		fmt.Fprintf(&b, "===== %s =====\n%s\n\n", f, strings.TrimSpace(string(data)))
	}
	return strings.TrimSpace(b.String()), nil
}

// ReadExerciseSet returns the markdown of one ingested exercise set
// (exercises/<course>/<set>.md), or ErrNoMaterial if it does not exist.
func (w Workspace) ReadExerciseSet(course, set string) (string, error) {
	data, err := os.ReadFile(w.ExerciseSetPath(course, set))
	if err != nil {
		if os.IsNotExist(err) {
			return "", ErrNoMaterial
		}
		return "", err
	}
	return string(data), nil
}

// ExerciseSets lists the set names (filenames without .md) under
// exercises/<course>/, sorted. A missing course directory yields an empty list.
func (w Workspace) ExerciseSets(course string) ([]string, error) {
	entries, err := os.ReadDir(w.Path("exercises", course))
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var sets []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".md") {
			sets = append(sets, strings.TrimSuffix(e.Name(), ".md"))
		}
	}
	sort.Strings(sets)
	return sets, nil
}

// MaterialFromFiles concatenates specific markdown files (relative to
// courses/<name>/) for scoped generation (FR-055). Missing files are an error.
func (w Workspace) MaterialFromFiles(name string, files []string) (string, error) {
	dir := w.Path("courses", name)
	var b strings.Builder
	for _, f := range files {
		path := filepath.Join(dir, f)
		data, err := os.ReadFile(path)
		if err != nil {
			return "", fmt.Errorf("scope file %q: %w", f, err)
		}
		fmt.Fprintf(&b, "===== %s =====\n%s\n\n", f, strings.TrimSpace(string(data)))
	}
	if b.Len() == 0 {
		return "", ErrNoMaterial
	}
	return strings.TrimSpace(b.String()), nil
}

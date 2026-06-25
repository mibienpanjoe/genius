package workspace

import (
	"errors"
	"testing"
)

func TestExerciseSetReaders(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("GENIUS_HOME", dir)
	w, err := Open(Config{})
	if err != nil {
		t.Fatal(err)
	}

	mustMkdir(t, w.Path("exercises", "algebra"))
	mustWrite(t, w.Path("exercises", "algebra", "td2.md"), "# td2 body")
	mustWrite(t, w.Path("exercises", "algebra", "td1.md"), "# td1 body")

	sets, err := w.ExerciseSets("algebra")
	if err != nil {
		t.Fatal(err)
	}
	if len(sets) != 2 || sets[0] != "td1" || sets[1] != "td2" {
		t.Errorf("ExerciseSets sorted names wrong: %v", sets)
	}

	body, err := w.ReadExerciseSet("algebra", "td1")
	if err != nil || body != "# td1 body" {
		t.Errorf("ReadExerciseSet: body=%q err=%v", body, err)
	}

	if _, err := w.ReadExerciseSet("algebra", "missing"); !errors.Is(err, ErrNoMaterial) {
		t.Errorf("missing set: want ErrNoMaterial, got %v", err)
	}

	// A course with no exercises directory yields an empty list, not an error.
	if sets, err := w.ExerciseSets("history"); err != nil || len(sets) != 0 {
		t.Errorf("absent course: want empty, got %v err=%v", sets, err)
	}
}

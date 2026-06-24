package workspace

import (
	"errors"
	"os"
	"testing"
)

func TestSlug(t *testing.T) {
	cases := map[string]string{
		"BIT-Course-For Student-Logic and Logic Programming.pdf": "bit-course-for-student-logic-and-logic-programming",
		"TD_BIT-2025-2026.pdf": "td-bit-2025-2026",
		"algebra.PPTX":         "algebra",
		"  ??? .pdf":           "untitled",
	}
	for in, want := range cases {
		if got := Slug(in); got != want {
			t.Errorf("Slug(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestWriteArtifactOverwriteGuard(t *testing.T) {
	dir := t.TempDir()
	w := Workspace{Root: dir}
	path := w.CourseDocPath("algebra", "lecture")

	if err := w.WriteArtifact(path, []byte("v1"), false); err != nil {
		t.Fatalf("first write: %v", err)
	}
	// Second write without force must refuse.
	if err := w.WriteArtifact(path, []byte("v2"), false); !errors.Is(err, ErrExists) {
		t.Fatalf("want ErrExists, got %v", err)
	}
	if b, _ := os.ReadFile(path); string(b) != "v1" {
		t.Errorf("file changed despite refusal: %q", b)
	}
	// Force overwrites.
	if err := w.WriteArtifact(path, []byte("v2"), true); err != nil {
		t.Fatalf("force write: %v", err)
	}
	if b, _ := os.ReadFile(path); string(b) != "v2" {
		t.Errorf("force did not overwrite: %q", b)
	}
}

package convert

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/mibienpanjoe/genius/internal/engine"
)

func TestNotationWarning(t *testing.T) {
	// Boolean material with bars dropped → warn.
	dropped := "Exercice 3\nF =XY+YZ+XZ\nKarnaugh maps"
	if notationWarning(dropped) == "" {
		t.Error("expected a notation warning when bars are missing from Boolean material")
	}
	// Bars present (LaTeX) → no warning.
	ok := `De Morgan: $\overline{a.b} = \overline{a}+\overline{b}$`
	if notationWarning(ok) != "" {
		t.Error("no warning expected when complement bars survive")
	}
	// Non-Boolean prose → no warning.
	if notationWarning("The mitochondria is the powerhouse of the cell.") != "" {
		t.Error("no warning expected for non-Boolean material")
	}
}

func TestFigureSectionPlaceholderOnNoVision(t *testing.T) {
	f := &engine.Fake{DescribeErr: engine.ErrNoVision}
	imgs := []ExtractedImage{{Path: "/w/assets/img-000.png", Page: 6, W: 148, H: 430}}
	section, paths := figureSection(context.Background(), imgs, f, true)

	if len(paths) != 1 {
		t.Fatalf("want 1 asset path, got %d", len(paths))
	}
	if !strings.Contains(section, "[Figure 1]") || !strings.Contains(section, "assets/img-000.png") {
		t.Errorf("figure section missing label/asset ref: %s", section)
	}
	if !strings.Contains(section, "image omitted") {
		t.Errorf("no-vision should leave a placeholder, not drop (FRB-09): %s", section)
	}
}

func TestEscalateNotationNoVision(t *testing.T) {
	// A vision-less engine (claude) cannot recover bars → empty, caller keeps
	// the warn-only banner.
	f := &engine.Fake{DescribeErr: engine.ErrNoVision}
	if got := escalateNotation(context.Background(), "/no/such.pdf", t.TempDir(), f); got != "" {
		t.Errorf("ErrNoVision must yield no correction, got: %q", got)
	}
}

func TestEscalateNotationNilEngine(t *testing.T) {
	if got := escalateNotation(context.Background(), "/no/such.pdf", t.TempDir(), nil); got != "" {
		t.Errorf("nil engine must yield no correction, got: %q", got)
	}
}

func TestEscalateNotationAssembly(t *testing.T) {
	const sample = "../../samples/TD_BIT-2025-2026.pdf"
	if !pagesAvailable() {
		t.Skip("pdftoppm not on PATH")
	}
	if _, err := os.Stat(sample); err != nil {
		t.Skip("sample PDF not present")
	}
	f := &engine.Fake{DescribeReply: "F = ĀBC + AB̄C + ABC̄"}
	got := escalateNotation(context.Background(), sample, t.TempDir(), f)
	if !strings.Contains(got, "Notation-corrected transcription") {
		t.Errorf("missing section heading: %s", got)
	}
	if !strings.Contains(got, "Page 1") || !strings.Contains(got, "Page 2") {
		t.Errorf("expected both pages transcribed (sample is 2pp): %s", got)
	}
	if !strings.Contains(got, "ĀBC") {
		t.Errorf("vision transcript not spliced in: %s", got)
	}
}

func TestFigureSectionCaption(t *testing.T) {
	f := &engine.Fake{DescribeReply: "A two-input AND gate with output Q = A∧B."}
	imgs := []ExtractedImage{{Path: "/w/assets/img-000.png", Page: 2, W: 300, H: 200}}
	section, _ := figureSection(context.Background(), imgs, f, true)

	if !strings.Contains(section, "AND gate") {
		t.Errorf("caption not inserted: %s", section)
	}
	if !strings.Contains(section, "transcribed from") {
		t.Errorf("provenance marker missing (INV-12): %s", section)
	}
}

package convert

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Two documents ingested into one shared assets dir must keep their figures
// apart: the second extraction must neither overwrite nor delete the first
// document's assets (they used to share the fixed "img" prefix).
func TestExtractImagesNoCrossDocumentCollision(t *testing.T) {
	if !imagesAvailable() {
		t.Skip("pdfimages not on PATH")
	}
	sample := ""
	for _, s := range []string{
		"../../samples/BIT-Course-For Student-Logic and Logic Programming.pdf",
		"../../samples/TD_BIT-2025-2026.pdf",
	} {
		if _, err := os.Stat(s); err == nil {
			sample = s
			break
		}
	}
	if sample == "" {
		t.Skip("no sample PDF present")
	}
	dir := t.TempDir()
	ctx := context.Background()

	// Same content under two names stands in for two course documents.
	second := filepath.Join(t.TempDir(), "chap02.pdf")
	data, err := os.ReadFile(sample)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(second, data, 0o644); err != nil {
		t.Fatal(err)
	}

	first, err := extractImages(ctx, sample, dir, 1)
	if err != nil {
		t.Fatal(err)
	}
	if len(first) == 0 {
		t.Skip("sample has no extractable images")
	}

	got, err := extractImages(ctx, second, dir, 1)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != len(first) {
		t.Errorf("identical content should yield the same count: first=%d second=%d",
			len(first), len(got))
	}

	// Every first-document asset must still exist after the second ingest.
	for _, im := range first {
		if _, err := os.Stat(im.Path); err != nil {
			t.Errorf("first document's asset lost after second ingest: %s", im.Path)
		}
	}
	// And the two documents' files must be distinguishable by prefix.
	for _, im := range got {
		base := filepath.Base(im.Path)
		if !strings.HasPrefix(base, "chap02-img-") {
			t.Errorf("second document's asset lacks its own prefix: %s", base)
		}
	}
}

// Re-ingesting the same document must not leave stale higher-numbered files
// behind (they would be misattributed to zero metadata and deleted, or worse,
// referenced by the markdown while gone).
func TestExtractImagesReingestCleansStale(t *testing.T) {
	if !imagesAvailable() {
		t.Skip("pdfimages not on PATH")
	}
	const sample = "../../samples/TD_BIT-2025-2026.pdf"
	if _, err := os.Stat(sample); err != nil {
		t.Skip("sample PDF not present")
	}
	dir := t.TempDir()

	// Plant a fake stale leftover from a previous, larger extraction.
	stale := filepath.Join(dir, docSlug(sample)+"-img-999.png")
	if err := os.WriteFile(stale, []byte("stale"), 0o644); err != nil {
		t.Fatal(err)
	}

	imgs, err := extractImages(context.Background(), sample, dir, 1)
	if err != nil {
		t.Fatal(err)
	}
	for _, im := range imgs {
		if filepath.Base(im.Path) == filepath.Base(stale) {
			t.Errorf("stale file misattributed to this run: %s", im.Path)
		}
	}
	if _, err := os.Stat(stale); err == nil {
		t.Errorf("stale leftover should have been cleared before extraction")
	}
}

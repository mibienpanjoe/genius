package convert

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
)

// pagesAvailable reports whether poppler's pdftoppm is on PATH (needed to
// rasterize whole pages for notation-fidelity escalation).
func pagesAvailable() bool {
	_, err := exec.LookPath("pdftoppm")
	return err == nil
}

// renderPages rasterizes every page of a PDF to PNGs in dir and returns the
// page image paths in page order (FR-035f). Used only on the escalation path,
// so the 150 DPI cost is paid solely for documents whose text extraction looks
// like it dropped notation.
func renderPages(ctx context.Context, pdf, dir string) ([]string, error) {
	if !pagesAvailable() {
		return nil, fmt.Errorf("pdftoppm (poppler) not found — cannot rasterize pages")
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, err
	}
	prefix := filepath.Join(dir, docSlug(pdf)+"-page")
	clearStale(prefix)
	cmd := exec.CommandContext(ctx, "pdftoppm", "-png", "-r", "150", pdf, prefix)
	if out, err := cmd.CombinedOutput(); err != nil {
		return nil, fmt.Errorf("pdftoppm: %w: %s", err, strings.TrimSpace(string(out)))
	}
	files, _ := filepath.Glob(prefix + "-*.png")
	sort.Strings(files) // pdftoppm zero-pads page numbers, so lexical == page order
	if len(files) == 0 {
		return nil, fmt.Errorf("pdftoppm produced no pages")
	}
	return files, nil
}

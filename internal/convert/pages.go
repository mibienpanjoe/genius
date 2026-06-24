package convert

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

// pagesAvailable reports whether poppler's pdftoppm is on PATH (needed to
// rasterize whole pages for notation-fidelity escalation).
func pagesAvailable() bool {
	_, err := exec.LookPath("pdftoppm")
	return err == nil
}

// pageCount returns the number of pages in a PDF via `pdfinfo`. A zero count
// with a nil error means pdfinfo ran but reported no "Pages:" line.
func pageCount(ctx context.Context, pdf string) (int, error) {
	if _, err := exec.LookPath("pdfinfo"); err != nil {
		return 0, fmt.Errorf("pdfinfo (poppler) not found")
	}
	out, err := exec.CommandContext(ctx, "pdfinfo", pdf).Output()
	if err != nil {
		return 0, fmt.Errorf("pdfinfo: %w", err)
	}
	sc := bufio.NewScanner(strings.NewReader(string(out)))
	for sc.Scan() {
		fields := strings.Fields(sc.Text())
		if len(fields) == 2 && fields[0] == "Pages:" {
			n, _ := strconv.Atoi(fields[1])
			return n, nil
		}
	}
	return 0, nil
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
	prefix := filepath.Join(dir, "page")
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

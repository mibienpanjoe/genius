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

// ExtractedImage is a raster figure pulled from a source document.
type ExtractedImage struct {
	Path string // saved asset path
	Page int    // 1-based source page
	W, H int    // pixel dimensions
}

// imagesAvailable reports whether poppler's pdfimages is on PATH.
func imagesAvailable() bool {
	_, err := exec.LookPath("pdfimages")
	return err == nil
}

// extractImages pulls embedded raster images from a PDF into dir as PNGs,
// skipping images whose width and height are both below minPx (decorative
// bullets/logos). Returns the kept images with page/size metadata (FR-035a).
func extractImages(ctx context.Context, pdf, dir string, minPx int) ([]ExtractedImage, error) {
	if !imagesAvailable() {
		return nil, fmt.Errorf("pdfimages (poppler) not found — cannot extract figures")
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, err
	}

	meta, err := listImages(ctx, pdf)
	if err != nil {
		return nil, err
	}

	prefix := filepath.Join(dir, "img")
	cmd := exec.CommandContext(ctx, "pdfimages", "-png", pdf, prefix)
	if out, err := cmd.CombinedOutput(); err != nil {
		return nil, fmt.Errorf("pdfimages extract: %w: %s", err, strings.TrimSpace(string(out)))
	}

	// pdfimages writes <prefix>-NNN.png in image-number order, matching -list.
	files, _ := filepath.Glob(prefix + "-*.png")
	sort.Strings(files)

	var kept []ExtractedImage
	for i, f := range files {
		var m imgMeta
		if i < len(meta) {
			m = meta[i]
		}
		// Soft masks / stencils are alpha channels, not figures — drop them.
		if m.typ != "" && m.typ != "image" {
			os.Remove(f)
			continue
		}
		if m.w < minPx && m.h < minPx {
			os.Remove(f) // decorative — drop the file too
			continue
		}
		kept = append(kept, ExtractedImage{Path: f, Page: m.page, W: m.w, H: m.h})
	}
	return kept, nil
}

type imgMeta struct {
	page, num, w, h int
	typ             string
}

// listImages parses `pdfimages -list` for per-image page and pixel size.
func listImages(ctx context.Context, pdf string) ([]imgMeta, error) {
	cmd := exec.CommandContext(ctx, "pdfimages", "-list", pdf)
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("pdfimages -list: %w", err)
	}
	var metas []imgMeta
	sc := bufio.NewScanner(strings.NewReader(string(out)))
	for sc.Scan() {
		fields := strings.Fields(sc.Text())
		// header rows: "page num type width height ..." and a dashes line.
		if len(fields) < 5 {
			continue
		}
		page, e1 := strconv.Atoi(fields[0])
		num, e2 := strconv.Atoi(fields[1])
		w, e3 := strconv.Atoi(fields[3])
		h, e4 := strconv.Atoi(fields[4])
		if e1 != nil || e2 != nil || e3 != nil || e4 != nil {
			continue
		}
		metas = append(metas, imgMeta{page: page, num: num, w: w, h: h, typ: fields[2]})
	}
	return metas, nil
}

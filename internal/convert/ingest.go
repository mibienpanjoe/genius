package convert

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"regexp"
	"strings"

	"genius/internal/engine"
)

// IngestOpts tunes a fidelity-aware ingest (FR-035).
type IngestOpts struct {
	Describe   bool // vision-caption extracted figures (default true)
	OCR        bool // attempt OCR/vision on text-less pages (opt-in)
	MinImagePx int  // skip images smaller than this in both dimensions
}

// DefaultMinImagePx filters out decorative bullets/logos.
const DefaultMinImagePx = 64

// Result is the outcome of an ingest: the markdown plus the saved asset paths.
type Result struct {
	Markdown string
	Assets   []string
}

// describeInstruction guides figure transcription with notation fidelity
// (INV-12): describe only what is present, preserve logic notation.
const describeInstruction = "Transcribe and briefly describe this figure from a " +
	"logic/computer-science course (it may be a circuit diagram, truth table, or " +
	"Karnaugh map). Reproduce any Boolean notation exactly, preserving complement " +
	"bars (write X̄), quantifiers (∀ ∃), and connectives (¬ ∧ ∨ → ↔). Describe ONLY " +
	"what is visibly present; do not invent content. Keep it under 80 words."

// Ingest converts the document to markdown and, for PDFs, extracts embedded
// figures into assetsDir, captioning them via the Engine (or leaving a
// provenance-marked placeholder when vision is unavailable). It also flags
// notation-dense pages whose text layer likely lost complement bars (INV-12).
func Ingest(ctx context.Context, path, assetsDir string, eng engine.Engine, opts IngestOpts) (Result, error) {
	if !Available() {
		return Result{}, ErrNotInstalled
	}
	md, err := ToMarkdown(ctx, path)
	if err != nil {
		return Result{}, err
	}

	minPx := opts.MinImagePx
	if minPx <= 0 {
		minPx = DefaultMinImagePx
	}

	isPDF := strings.EqualFold(filepath.Ext(path), ".pdf")

	var assets []string
	if isPDF && imagesAvailable() {
		imgs, ierr := extractImages(ctx, path, assetsDir, minPx)
		if ierr == nil && len(imgs) > 0 {
			section, paths := figureSection(ctx, imgs, assetsDir, eng, opts.Describe)
			md += "\n\n" + section
			assets = paths
		}
	}

	if banner := notationWarning(md); banner != "" {
		// Escalate-on-detect: the text layer looks like it dropped complement
		// bars. For PDFs with a vision engine, rasterize the pages and have the
		// model re-transcribe them, recovering notation that lives only in the
		// pixels (INV-12/FR-035f). Cost is paid solely for flagged documents.
		corrected := ""
		if isPDF && opts.Describe {
			corrected = escalateNotation(ctx, path, assetsDir, eng)
		}
		if corrected != "" {
			md = notationCorrectedBanner + "\n\n" + md + "\n\n" + corrected
		} else {
			md = banner + "\n\n" + md
		}
	}

	return Result{Markdown: md, Assets: assets}, nil
}

// figureSection builds the "Extracted Figures" markdown and returns it with the
// list of asset paths. Each figure gets a provenance-marked blockquote
// (FR-035b/c, INV-12): a model caption when describable, else a placeholder.
func figureSection(ctx context.Context, imgs []ExtractedImage, assetsDir string, eng engine.Engine, describe bool) (string, []string) {
	var b strings.Builder
	b.WriteString("## Extracted Figures\n")
	b.WriteString("_Model transcriptions of figures pulled from the source. " +
		"Verify against the original._\n\n")

	var paths []string
	for i, im := range imgs {
		rel := relAsset(assetsDir, im.Path)
		paths = append(paths, im.Path)

		caption := ""
		if describe && eng != nil {
			out, err := eng.Describe(ctx, im.Path, describeInstruction)
			if err == nil && strings.TrimSpace(out) != "" {
				caption = strings.TrimSpace(out)
			}
		}
		if caption == "" {
			// EXC-12: no vision — placeholder, never a silent drop (FRB-09).
			fmt.Fprintf(&b, "> **[Figure %d]** _image omitted_ (transcribe unavailable) "+
				"— %s, page %d\n\n", i+1, rel, im.Page)
			continue
		}
		fmt.Fprintf(&b, "> **[Figure %d]** %s\n>\n> _(transcribed from %s — page %d)_\n\n",
			i+1, caption, rel, im.Page)
	}
	return b.String(), paths
}

func relAsset(assetsDir, p string) string {
	return "assets/" + filepath.Base(p)
}

// pageTranscribeInstruction drives full-page re-transcription on the escalation
// path. Unlike describeInstruction (figures, terse), this asks for a faithful
// verbatim transcription of an entire page, with explicit emphasis on the
// complement bars the text layer tends to drop (INV-12).
const pageTranscribeInstruction = "Transcribe this full page from a logic/" +
	"computer-science course into markdown. Reproduce ALL text and Boolean " +
	"notation EXACTLY, preserving complement bars (write X̄ with the overline), " +
	"quantifiers (∀ ∃), negation and connectives (¬ ∧ ∨ → ↔), subscripts, and " +
	"truth tables. Pay special attention to overbars on Boolean variables — they " +
	"are easy to miss and inverting them changes the meaning. Transcribe ONLY " +
	"what is visibly present; do not solve, explain, or invent. Preserve the " +
	"original reading order and structure."

// notationCorrectedBanner replaces the warn-only banner once a vision
// re-transcription has been appended, pointing the reader (and the generation
// model) at the authoritative section.
const notationCorrectedBanner = "> ⚠ **Notation check (INV-12):** the plain-text " +
	"extraction below looks like it **dropped Boolean complement bars** (X̄). A " +
	"vision re-transcription was appended under _Notation-corrected transcription_ " +
	"— treat that section as authoritative for any Boolean/logic notation."

// escalateNotation rasterizes a PDF's pages and re-transcribes each via the
// engine's vision path, returning a markdown section that recovers notation the
// text layer dropped. Returns "" on any obstacle (no vision, render failure, no
// usable output) so the caller falls back to the warn-only banner.
func escalateNotation(ctx context.Context, pdf, assetsDir string, eng engine.Engine) string {
	if eng == nil || !pagesAvailable() {
		return ""
	}
	pages, err := renderPages(ctx, pdf, assetsDir)
	if err != nil || len(pages) == 0 {
		return ""
	}

	var b strings.Builder
	b.WriteString("## Notation-corrected transcription (vision)\n")
	b.WriteString("_Authoritative for Boolean/logic notation: the plain-text " +
		"extraction above dropped complement bars. Verify against the original._\n\n")

	got := 0
	for i, p := range pages {
		out, derr := eng.Describe(ctx, p, pageTranscribeInstruction)
		if errors.Is(derr, engine.ErrNoVision) {
			return "" // engine has no vision at all — nothing to recover
		}
		if derr != nil || strings.TrimSpace(out) == "" {
			continue
		}
		fmt.Fprintf(&b, "### Page %d (transcribed from %s)\n\n%s\n\n",
			i+1, relAsset(assetsDir, p), strings.TrimSpace(out))
		got++
	}
	if got == 0 {
		return ""
	}
	return b.String()
}

// boolFuncRe spots Boolean-function definitions like "F =XY+YZ" that, after a
// lossy text extraction, may have lost their complement bars.
var boolFuncRe = regexp.MustCompile(`\b[A-Z]\s*=\s*[A-Z][A-Z0-9 +().]*`)

// hasOverline reports whether the text carries combining/standalone overline
// marks or LaTeX \overline (i.e. complement bars survived).
func hasOverline(s string) bool {
	if strings.Contains(s, "̄") || strings.Contains(s, "̅") ||
		strings.Contains(s, "\\overline") || strings.Contains(s, "‾") {
		return true
	}
	return false
}

// notationWarning returns a banner when the document looks like Boolean-logic
// material but carries no complement-bar notation — a strong sign the text
// extraction silently dropped the bars (INV-12/FRB-10/FR-035f).
func notationWarning(md string) string {
	if hasOverline(md) {
		return ""
	}
	low := strings.ToLower(md)
	boolish := strings.Contains(low, "karnaugh") ||
		strings.Contains(low, "fonction") && strings.Contains(low, "booléen") ||
		strings.Contains(low, "boolean") ||
		boolFuncRe.MatchString(md)
	if !boolish {
		return ""
	}
	return "> ⚠ **Notation check (INV-12):** this looks like Boolean-logic material " +
		"but no complement bars (X̄) were detected in the extracted text. Text " +
		"extraction may have **dropped complement bars**, which inverts meaning. " +
		"Re-ingest with vision transcription (`--ocr`/`--describe-images`) or verify " +
		"Boolean functions against the original before relying on them."
}

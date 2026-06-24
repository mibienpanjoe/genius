package tui

// wordmark is the block-letter "genius" banner (toilet `pagga` font), colored
// per-column with the brand gradient at render time. See docs/07_visual_identity.md.
const wordmark = "" +
	"░█▀▀░█▀▀░█▀█░▀█▀░█░█░█▀▀\n" +
	"░█░█░█▀▀░█░█░░█░░█░█░▀▀█\n" +
	"░▀▀▀░▀▀▀░▀░▀░▀▀▀░▀▀▀░▀▀▀"

// wordmarkWidth is the column count of the banner (all rows equal width).
const wordmarkWidth = 23

// renderWordmark returns the gradient-colored banner. Falls back to a compact
// single-line gradient "genius" for narrow terminals.
func renderWordmark(termWidth int) string {
	if termWidth < wordmarkWidth+4 {
		return gradientLine("genius", 6)
	}
	return gradientBlock(wordmark)
}

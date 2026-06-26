package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Palette — tokens from docs/07_visual_identity.md. Truecolor hex; Lip Gloss
// degrades to 256/16-color automatically. NO_COLOR is honored by Lip Gloss.
var (
	cPrimary   = lipgloss.Color("#7C6FF0") // Genius Violet
	cSurface   = lipgloss.Color("#1A1A24") // Panel
	cText      = lipgloss.Color("#E6E6F0") // Paper
	cTextMuted = lipgloss.Color("#8A8AA0") // Slate

	cSuccess = lipgloss.Color("#3FB950")
	cWarning = lipgloss.Color("#D6A33E")
	cError   = lipgloss.Color("#E5534B")
	cInfo    = lipgloss.Color("#58A6FF")

	cLine = lipgloss.Color("#2A2A38")
)

// Brand gradient stops (sky → violet → orchid → pink).
var gradientStops = []struct {
	pos float64
	r   int
	g   int
	b   int
}{
	{0.00, 0x6B, 0xA8, 0xF5},
	{0.35, 0x7C, 0x6F, 0xF0},
	{0.70, 0xB0, 0x6F, 0xE6},
	{1.00, 0xE6, 0x6F, 0xB0},
}

// Shared styles.
var (
	styleTitle  = lipgloss.NewStyle().Bold(true).Foreground(cPrimary)
	styleBody   = lipgloss.NewStyle().Foreground(cText)
	styleMuted  = lipgloss.NewStyle().Foreground(cTextMuted)
	styleInfo   = lipgloss.NewStyle().Foreground(cInfo)
	styleKey    = lipgloss.NewStyle().Foreground(cPrimary)
	styleStatus = lipgloss.NewStyle().Background(cSurface).Foreground(cTextMuted)

	styleNotice = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(cWarning).
			Foreground(cText).
			Padding(0, 2)
)

// gradientColorAt returns the interpolated brand color at fraction t in [0,1].
func gradientColorAt(t float64) lipgloss.Color {
	if t <= 0 {
		return rgbHex(gradientStops[0].r, gradientStops[0].g, gradientStops[0].b)
	}
	if t >= 1 {
		last := gradientStops[len(gradientStops)-1]
		return rgbHex(last.r, last.g, last.b)
	}
	for i := 1; i < len(gradientStops); i++ {
		hi := gradientStops[i]
		if t <= hi.pos {
			lo := gradientStops[i-1]
			span := hi.pos - lo.pos
			f := 0.0
			if span > 0 {
				f = (t - lo.pos) / span
			}
			return rgbHex(
				lerp(lo.r, hi.r, f),
				lerp(lo.g, hi.g, f),
				lerp(lo.b, hi.b, f),
			)
		}
	}
	last := gradientStops[len(gradientStops)-1]
	return rgbHex(last.r, last.g, last.b)
}

func lerp(a, b int, f float64) int {
	return a + int(float64(b-a)*f+0.5)
}

func rgbHex(r, g, b int) lipgloss.Color {
	const hexdig = "0123456789ABCDEF"
	buf := []byte{'#', 0, 0, 0, 0, 0, 0}
	vals := []int{r, g, b}
	for i, v := range vals {
		buf[1+i*2] = hexdig[(v>>4)&0xF]
		buf[2+i*2] = hexdig[v&0xF]
	}
	return lipgloss.Color(string(buf))
}

// gradientLine colors a single line of text per-cell across the brand gradient.
func gradientLine(s string, width int) string {
	runes := []rune(s)
	if len(runes) == 0 {
		return s
	}
	denom := width
	if denom <= 1 {
		denom = len(runes)
	}
	var b strings.Builder
	for i, r := range runes {
		if r == ' ' {
			b.WriteRune(r)
			continue
		}
		t := 0.0
		if denom > 1 {
			t = float64(i) / float64(denom-1)
		}
		b.WriteString(lipgloss.NewStyle().Foreground(gradientColorAt(t)).Render(string(r)))
	}
	return b.String()
}

// gradientBlock applies the gradient per-column across a multi-line block, so
// every row shares the same left-to-right ramp (the wordmark banner).
func gradientBlock(block string) string {
	lines := strings.Split(block, "\n")
	width := 0
	for _, ln := range lines {
		if w := len([]rune(ln)); w > width {
			width = w
		}
	}
	out := make([]string, len(lines))
	for i, ln := range lines {
		runes := []rune(ln)
		var b strings.Builder
		for j, r := range runes {
			if r == ' ' {
				b.WriteRune(r)
				continue
			}
			t := 0.0
			if width > 1 {
				t = float64(j) / float64(width-1)
			}
			b.WriteString(lipgloss.NewStyle().Foreground(gradientColorAt(t)).Render(string(r)))
		}
		out[i] = b.String()
	}
	return strings.Join(out, "\n")
}

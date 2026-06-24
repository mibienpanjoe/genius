// Package render turns markdown into styled terminal output in-process using
// Glamour (no external pager), with a Genius-branded style (FR-071, INV §07).
package render

import (
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/glamour/ansi"
	"github.com/charmbracelet/glamour/styles"
)

const (
	violet = "#7C6FF0" // Genius Violet — headings
	info   = "#58A6FF" // links
)

// geniusStyle derives from Glamour's dark style and recolors headings/links to
// the brand palette (docs/07 §Rendered markdown).
func geniusStyle() ansi.StyleConfig {
	s := styles.DarkStyleConfig
	v := violet
	i := info
	s.H1.Color = &v
	s.H2.Color = &v
	s.H3.Color = &v
	s.H4.Color = &v
	s.Link.Color = &i
	return s
}

// Markdown renders md to a styled string wrapped at the given width. A width
// of <=0 disables wrapping.
func Markdown(md string, width int) (string, error) {
	opts := []glamour.TermRendererOption{glamour.WithStyles(geniusStyle())}
	if width > 0 {
		opts = append(opts, glamour.WithWordWrap(width))
	}
	r, err := glamour.NewTermRenderer(opts...)
	if err != nil {
		return "", err
	}
	return r.Render(md)
}

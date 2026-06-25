package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// helpKeys is the full keybinding reference shown by the `?` overlay, grouped by
// where each key applies. docs/07: every action has a discoverable key.
var helpKeys = []struct {
	group string
	rows  [][2]string // {keys, description}
}{
	{"home", [][2]string{
		{"↑/↓ · j/k", "move between courses"},
		{"g · enter", "open guide (build it if missing)"},
		{"q", "open q&a (build it if missing)"},
		{"G / Q", "force-regenerate guide / q&a"},
		{"r", "revise (quiz from q&a)"},
		{"s", "solve exercises"},
		{"i", "ingest a document"},
		{"f", "scope guide/q&a to chapters"},
	}},
	{"ingest", [][2]string{
		{"↑/↓", "browse · enter open dir / pick file"},
		{"space", "select file (batch)"},
		{"⌫", "up a directory"},
		{"tab", "switch course / exercise kind"},
	}},
	{"reader", [][2]string{
		{"↑/↓ · pgup/pgdn", "scroll"},
		{"q · esc", "back"},
	}},
	{"quiz", [][2]string{
		{"enter", "reveal answer"},
		{"y / n", "grade: knew it / missed"},
		{"space", "next question"},
		{"esc", "back"},
	}},
	{"solve", [][2]string{
		{"↑/↓", "move"},
		{"space", "toggle exercise"},
		{"enter", "solve selected (or current)"},
		{"esc", "back"},
	}},
	{"global", [][2]string{
		{"?", "this help"},
		{"ctrl+c", "quit"},
	}},
}

// viewHelp renders the keybinding reference as a centered card.
func (m Model) viewHelp() string {
	var b strings.Builder
	b.WriteString(styleTitle.Render("keys"))
	b.WriteString("\n")

	for _, g := range helpKeys {
		b.WriteString("\n")
		b.WriteString(lipgloss.NewStyle().Bold(true).Foreground(cText).Render(g.group))
		b.WriteString("\n")
		for _, kv := range g.rows {
			b.WriteString(styleKey.Render(fitCell(kv[0], 18)))
			b.WriteString(styleMuted.Render(kv[1]))
			b.WriteString("\n")
		}
	}

	b.WriteString("\n")
	b.WriteString(styleMuted.Render("any key to close"))

	card := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(cPrimary).
		Padding(1, 2).
		Render(b.String())

	return center(m.width, m.contentHeight(), card)
}

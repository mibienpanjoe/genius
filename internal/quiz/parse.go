// Package quiz parses revision Q&A files into ordered question/answer pairs for
// the interactive revise loop (FR-061).
package quiz

import (
	"fmt"
	"regexp"
	"strings"
)

// Pair is one revision question and its model answer.
type Pair struct {
	N        int
	Question string
	Answer   string
}

// headingRe matches a question heading: `## Q<n>. <question>` (FR-052).
var headingRe = regexp.MustCompile(`^##\s*Q(\d+)\.\s*(.*)$`)

// Parse reads a Q&A markdown file into ordered pairs. Content before the first
// question heading (title/intro) and a trailing `---` footer (motivational
// closing) are ignored. ErrNoQuestions is returned when nothing parses
// (ERR-061).
func Parse(md string) ([]Pair, error) {
	lines := strings.Split(md, "\n")
	var pairs []Pair
	var cur *Pair
	var body []string

	flush := func() {
		if cur != nil {
			cur.Answer = strings.TrimSpace(strings.Join(body, "\n"))
			pairs = append(pairs, *cur)
		}
		cur = nil
		body = nil
	}

	for _, ln := range lines {
		if m := headingRe.FindStringSubmatch(strings.TrimSpace(ln)); m != nil {
			flush()
			n := 0
			fmt.Sscanf(m[1], "%d", &n)
			cur = &Pair{N: n, Question: strings.TrimSpace(m[2])}
			continue
		}
		// A horizontal rule after the last question ends the Q&A body
		// (footer follows); ignore everything from there.
		if cur != nil && strings.TrimSpace(ln) == "---" {
			flush()
			break
		}
		if cur != nil {
			body = append(body, ln)
		}
	}
	flush()

	if len(pairs) == 0 {
		return nil, ErrNoQuestions
	}
	return pairs, nil
}

// ErrNoQuestions signals a Q&A file with no parseable `## Q<n>.` headings.
var ErrNoQuestions = fmt.Errorf("no question headings found (expected '## Q<n>. …')")

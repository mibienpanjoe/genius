package generate

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
)

// Exercise is one addressable problem within a set, with optional sub-questions
// (docs/06). Sub-parts are addressed as "<num>.<marker>", e.g. "3.1" or "3.a"
// (FR-103).
type Exercise struct {
	Label string     // human label, e.g. "Exercice 3" / "Problem 2"
	Num   string     // addressable identifier extracted from the label, e.g. "3"
	Text  string     // the full statement under the heading (incl. sub-parts)
	Parts []Exercise // parsed sub-questions, addressable via Num ("1", "a")
}

// ErrNoExercises signals a set with no recognizable exercise headings (ERR-101);
// the caller may fall back to solving the whole set as one item.
var ErrNoExercises = errors.New("no exercises found in set")

// ErrUnknownExercise signals a requested label that the set does not contain
// (ERR-102).
var ErrUnknownExercise = errors.New("unknown exercise")

// exHeadRe matches an exercise heading in FR or EN, optionally wrapped in
// markdown heading hashes or bold stars: "## Exercice 3", "**Exercise 2:**",
// "Problème 1 —". The number/label is captured for addressing.
var exHeadRe = regexp.MustCompile(
	`(?i)^\s*(?:#+\s*)?\*{0,2}\s*(exercice|exercise|problème|probleme|problem|ex)\b\s*[:.\-–—]?\s*([0-9]+|[ivxlc]+|[a-z])\b`)

// subHeadRe matches a sub-question marker at the start of a line: "1.", "2)",
// "(a)", "a)". The first non-empty group is the marker.
var subHeadRe = regexp.MustCompile(
	`^\s*(?:\(([0-9a-zA-Z]+)\)|([0-9]+)[.)]|([a-zA-Z])[.)])\s+\S`)

// Enumerate parses a set's markdown into an ordered, addressable list of
// exercises with their sub-parts (FR-102). It returns ErrNoExercises when no
// heading is recognized (ERR-101).
func Enumerate(setMD string) ([]Exercise, error) {
	lines := strings.Split(setMD, "\n")

	var exs []Exercise
	var cur *Exercise
	var body []string

	flush := func() {
		if cur != nil {
			cur.Text = strings.TrimSpace(strings.Join(body, "\n"))
			cur.Parts = parseParts(body)
			exs = append(exs, *cur)
		}
		cur = nil
		body = nil
	}

	for _, ln := range lines {
		if m := exHeadRe.FindStringSubmatch(ln); m != nil {
			flush()
			label := titleWord(m[1]) + " " + m[2]
			cur = &Exercise{Label: label, Num: strings.ToLower(m[2])}
			body = append(body, ln)
			continue
		}
		if cur != nil {
			body = append(body, ln)
		}
	}
	flush()

	if len(exs) == 0 {
		return nil, ErrNoExercises
	}
	return exs, nil
}

// parseParts extracts sub-questions from an exercise body. The body's first line
// is the exercise heading itself, which is skipped.
func parseParts(body []string) []Exercise {
	var parts []Exercise
	var cur *Exercise
	var buf []string

	flush := func() {
		if cur != nil {
			cur.Text = strings.TrimSpace(strings.Join(buf, "\n"))
			parts = append(parts, *cur)
		}
		cur = nil
		buf = nil
	}

	for i, ln := range body {
		if i == 0 {
			continue // the exercise heading line
		}
		if m := subHeadRe.FindStringSubmatch(ln); m != nil {
			marker := firstNonEmpty(m[1], m[2], m[3])
			flush()
			cur = &Exercise{Label: marker, Num: strings.ToLower(marker)}
			buf = append(buf, ln)
			continue
		}
		if cur != nil {
			buf = append(buf, ln)
		}
	}
	flush()
	return parts
}

// Select resolves selection tokens (e.g. "2", "3.1", "3.a") against the
// enumerated exercises, returning the chosen items in request order. A sub-part
// token yields a synthetic Exercise carrying just that part's statement. An
// unmatched token returns ErrUnknownExercise (ERR-102).
func Select(exs []Exercise, tokens []string) ([]Exercise, error) {
	var out []Exercise
	for _, raw := range tokens {
		tok := strings.TrimSpace(raw)
		if tok == "" {
			continue
		}
		exNum, partNum, hasPart := strings.Cut(tok, ".")
		ex := findByNum(exs, exNum)
		if ex == nil {
			return nil, fmt.Errorf("%w: %s", ErrUnknownExercise, tok)
		}
		if !hasPart {
			out = append(out, *ex)
			continue
		}
		part := findByNum(ex.Parts, partNum)
		if part == nil {
			return nil, fmt.Errorf("%w: %s", ErrUnknownExercise, tok)
		}
		out = append(out, Exercise{
			Label: ex.Label + "." + part.Num,
			Num:   tok,
			Text:  part.Text,
		})
	}
	return out, nil
}

func findByNum(exs []Exercise, num string) *Exercise {
	num = strings.ToLower(strings.TrimSpace(num))
	for i := range exs {
		if exs[i].Num == num {
			return &exs[i]
		}
	}
	return nil
}

func firstNonEmpty(ss ...string) string {
	for _, s := range ss {
		if s != "" {
			return s
		}
	}
	return ""
}

// titleWord lowercases a word then capitalizes its first rune, normalizing
// heading keywords like "EXERCICE"/"exercise" to "Exercice"/"Exercise".
func titleWord(s string) string {
	s = strings.ToLower(s)
	if s == "" {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

package quiz

import (
	"errors"
	"strings"
	"testing"
)

func TestParseRealFormat(t *testing.T) {
	md := `# Mobile Networks — Q&A

## Q1. What is handover?

The process of moving a call between cells.

> Key rule: continuity matters.

## Q2. Define SNR.

Signal-to-noise ratio, $SNR = P_s/P_n$.

---

*Bon courage pour ton examen! 🎯*`

	pairs, err := Parse(md)
	if err != nil {
		t.Fatal(err)
	}
	if len(pairs) != 2 {
		t.Fatalf("want 2 pairs, got %d", len(pairs))
	}
	if pairs[0].N != 1 || !strings.HasPrefix(pairs[0].Question, "What is handover") {
		t.Errorf("Q1 wrong: %+v", pairs[0])
	}
	if !strings.Contains(pairs[0].Answer, "moving a call") || !strings.Contains(pairs[0].Answer, "Key rule") {
		t.Errorf("Q1 answer should include body + callout: %q", pairs[0].Answer)
	}
	// Footer after --- must be excluded from the last answer.
	if strings.Contains(pairs[1].Answer, "Bon courage") {
		t.Errorf("footer leaked into answer: %q", pairs[1].Answer)
	}
	if !strings.Contains(pairs[1].Answer, "SNR = P_s") {
		t.Errorf("Q2 answer missing formula: %q", pairs[1].Answer)
	}
}

func TestParseNoQuestions(t *testing.T) {
	if _, err := Parse("# just a title\n\nsome prose"); !errors.Is(err, ErrNoQuestions) {
		t.Errorf("want ErrNoQuestions, got %v", err)
	}
}

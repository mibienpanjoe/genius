package generate

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/mibienpanjoe/genius/internal/engine"
)

const sampleSet = `# TD 3 — Algèbre de Boole

Exercice 1
Simplifier la fonction F = AB + AB̄.

## Exercice 2 : Karnaugh
Soit la fonction suivante.
1. Dresser la table de vérité.
2. Tracer le tableau de Karnaugh.

**Problem 3**
Prove the absorption law.
(a) State it.
(b) Prove it.
`

func TestEnumerate(t *testing.T) {
	exs, err := Enumerate(sampleSet)
	if err != nil {
		t.Fatalf("Enumerate: %v", err)
	}
	if len(exs) != 3 {
		t.Fatalf("want 3 exercises, got %d: %+v", len(exs), exs)
	}
	if exs[0].Num != "1" || exs[1].Num != "2" || exs[2].Num != "3" {
		t.Errorf("numbers off: %q %q %q", exs[0].Num, exs[1].Num, exs[2].Num)
	}
	if exs[1].Label != "Exercice 2" {
		t.Errorf("label normalize: got %q", exs[1].Label)
	}
	// Exercise 2 has numbered sub-parts; Problem 3 has lettered ones.
	if len(exs[1].Parts) != 2 || exs[1].Parts[0].Num != "1" {
		t.Errorf("numbered sub-parts wrong: %+v", exs[1].Parts)
	}
	if len(exs[2].Parts) != 2 || exs[2].Parts[1].Num != "b" {
		t.Errorf("lettered sub-parts wrong: %+v", exs[2].Parts)
	}
	// Statement text is preserved (bar survives).
	if !strings.Contains(exs[0].Text, "AB̄") {
		t.Errorf("statement text lost: %q", exs[0].Text)
	}
}

func TestEnumerateNone(t *testing.T) {
	if _, err := Enumerate("Just some prose, no exercises here."); !errors.Is(err, ErrNoExercises) {
		t.Errorf("want ErrNoExercises, got %v", err)
	}
}

func TestSelect(t *testing.T) {
	exs, err := Enumerate(sampleSet)
	if err != nil {
		t.Fatal(err)
	}

	// Whole exercise by number.
	got, err := Select(exs, []string{"1"})
	if err != nil || len(got) != 1 || got[0].Num != "1" {
		t.Fatalf("select whole: %+v err=%v", got, err)
	}

	// Sub-part by number and by letter, order preserved.
	got, err = Select(exs, []string{"3.b", "2.1"})
	if err != nil {
		t.Fatalf("select parts: %v", err)
	}
	if len(got) != 2 || got[0].Label != "Problem 3.b" || got[1].Label != "Exercice 2.1" {
		t.Errorf("sub-part select wrong: %+v", got)
	}
	if !strings.Contains(got[1].Text, "table de vérité") {
		t.Errorf("sub-part text not isolated: %q", got[1].Text)
	}
}

func TestSelectUnknown(t *testing.T) {
	exs, _ := Enumerate(sampleSet)
	if _, err := Select(exs, []string{"9"}); !errors.Is(err, ErrUnknownExercise) {
		t.Errorf("unknown exercise: want ErrUnknownExercise, got %v", err)
	}
	if _, err := Select(exs, []string{"3.z"}); !errors.Is(err, ErrUnknownExercise) {
		t.Errorf("unknown sub-part: want ErrUnknownExercise, got %v", err)
	}
}

func TestSolveGrounded(t *testing.T) {
	exs, _ := Enumerate(sampleSet)
	sel, _ := Select(exs, []string{"1"})
	f := &engine.Fake{Reply: "## Exercice 1\n**Answer** — F = A."}

	out, err := Solve(context.Background(), f, "algebra", "Boolean course material.", sel)
	if err != nil {
		t.Fatalf("Solve: %v", err)
	}
	if !strings.Contains(out, "F = A") {
		t.Errorf("engine reply not returned: %q", out)
	}
	// Grounding and the exact statement reach the engine.
	if !strings.Contains(f.LastUser, "Boolean course material") {
		t.Errorf("material not in prompt: %q", f.LastUser)
	}
	if !strings.Contains(f.LastUser, "Exercice 1") {
		t.Errorf("exercise label not in prompt: %q", f.LastUser)
	}
}

func TestSolveRefusesWithoutMaterial(t *testing.T) {
	sel := []Exercise{{Label: "Exercice 1", Text: "..."}}
	f := &engine.Fake{Reply: "should not be called"}
	if _, err := Solve(context.Background(), f, "algebra", "   ", sel); err == nil {
		t.Error("expected refusal on empty material (INV-05)")
	}
	if f.LastUser != "" {
		t.Errorf("engine must not be called when grounding is empty, got prompt: %q", f.LastUser)
	}
}

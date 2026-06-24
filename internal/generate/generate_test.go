package generate

import (
	"context"
	"strings"
	"testing"

	"genius/internal/engine"
)

func TestGuideRefusesEmptyMaterial(t *testing.T) {
	f := &engine.Fake{Reply: "x"}
	if _, err := Guide(context.Background(), f, "algebra", "  "); err == nil {
		t.Error("guide should refuse empty material")
	}
	if f.Calls != 0 {
		t.Error("engine must not be called without material (INV-05)")
	}
}

func TestGuidePassesMaterialAndSystemPrompt(t *testing.T) {
	f := &engine.Fake{Reply: "# guide"}
	_, err := Guide(context.Background(), f, "logic", "complement X̄ and ∀x")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(f.LastUser, "complement X̄ and ∀x") {
		t.Error("material not passed to engine")
	}
	if !strings.Contains(f.LastSys, "complement bar") {
		t.Error("guide system prompt should carry notation rules")
	}
}

func TestQACountAndScope(t *testing.T) {
	f := &engine.Fake{Reply: "## Q1. x"}
	_, err := QA(context.Background(), f, "logic", "material", QAOpts{Count: 7, Scope: "Karnaugh maps"})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(f.LastUser, "exactly 7 Q&A") {
		t.Errorf("count not honored in prompt: %s", f.LastUser)
	}
	if !strings.Contains(f.LastUser, "Karnaugh maps") {
		t.Error("scope not passed")
	}
}

func TestQADefaultCount(t *testing.T) {
	f := &engine.Fake{Reply: "## Q1. x"}
	if _, err := QA(context.Background(), f, "c", "m", QAOpts{}); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(f.LastUser, "exactly 10 Q&A") {
		t.Errorf("default count not applied: %s", f.LastUser)
	}
}

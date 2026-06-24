package engine

import (
	"context"
	"errors"
	"testing"
)

func TestNewSelectsBackend(t *testing.T) {
	cases := map[string]string{"": "claude", "claude": "claude", "codex": "codex"}
	for in, want := range cases {
		e, err := New(in, "")
		if err != nil {
			t.Fatalf("New(%q): %v", in, err)
		}
		if e.Name() != want {
			t.Errorf("New(%q).Name() = %q, want %q", in, e.Name(), want)
		}
	}
	if _, err := New("bogus", ""); err == nil {
		t.Error("unknown engine should error")
	}
}

func TestFakeEngine(t *testing.T) {
	f := &Fake{Reply: "hello"}
	got, err := f.Generate(context.Background(), "sys", "user")
	if err != nil || got != "hello" {
		t.Fatalf("got %q, %v", got, err)
	}
	if f.LastSys != "sys" || f.LastUser != "user" || f.Calls != 1 {
		t.Errorf("fake did not record call: %+v", f)
	}

	want := errors.New("boom")
	f2 := &Fake{Err: want}
	if _, err := f2.Generate(context.Background(), "", ""); !errors.Is(err, want) {
		t.Errorf("fake should return its error")
	}
}

// Engine implementations satisfy the interface.
var _ Engine = (*claudeEngine)(nil)
var _ Engine = (*codexEngine)(nil)
var _ Engine = (*Fake)(nil)

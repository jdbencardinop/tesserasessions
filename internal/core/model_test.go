package core

import "testing"

func TestCanonicalIDStable(t *testing.T) {
	a := CanonicalID("claude", "abc")
	b := CanonicalID("claude", "abc")
	if a != b {
		t.Fatalf("expected stable id, got %q and %q", a, b)
	}
	if a == CanonicalID("copilot", "abc") {
		t.Fatalf("source should affect id")
	}
}

func TestShellQuote(t *testing.T) {
	got := ShellQuote("it's here")
	want := "'it'\\''s here'"
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}

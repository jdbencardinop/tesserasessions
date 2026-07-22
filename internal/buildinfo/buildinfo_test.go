package buildinfo

import (
	"runtime/debug"
	"testing"
)

func TestResolvePrefersInjectedVersion(t *testing.T) {
	got := resolve("v1.2.3", func() (*debug.BuildInfo, bool) {
		return &debug.BuildInfo{Main: debug.Module{Version: "v9.9.9"}}, true
	})
	if got != "v1.2.3" {
		t.Fatalf("expected injected version to win, got %q", got)
	}
}

func TestResolveUsesModuleVersionWhenInjectedIsDev(t *testing.T) {
	got := resolve("dev", func() (*debug.BuildInfo, bool) {
		return &debug.BuildInfo{Main: debug.Module{Version: "v0.1.0"}}, true
	})
	if got != "v0.1.0" {
		t.Fatalf("expected module version, got %q", got)
	}
}

func TestResolveKeepsDevForDevelBuild(t *testing.T) {
	got := resolve("dev", func() (*debug.BuildInfo, bool) {
		return &debug.BuildInfo{Main: debug.Module{Version: "(devel)"}}, true
	})
	if got != "dev" {
		t.Fatalf("expected dev fallback, got %q", got)
	}
}

func TestResolveKeepsDevWithoutBuildInfo(t *testing.T) {
	got := resolve("dev", func() (*debug.BuildInfo, bool) {
		return nil, false
	})
	if got != "dev" {
		t.Fatalf("expected dev fallback, got %q", got)
	}
}

func TestStringDoesNotPanic(t *testing.T) {
	if got := String(); got == "" {
		t.Fatal("String returned empty")
	}
}

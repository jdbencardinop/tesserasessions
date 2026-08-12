package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultHermesDatabaseHonorsHomeAndActiveProfile(t *testing.T) {
	home := t.TempDir()
	customRoot := filepath.Join(home, "custom")
	if err := os.MkdirAll(customRoot, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(customRoot, "active_profile"), []byte("coder\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("HERMES_HOME", customRoot)
	if got, want := defaultHermesDatabase(home), filepath.Join(customRoot, "profiles", "coder", "state.db"); got != want {
		t.Fatalf("root HERMES_HOME path = %q, want %q", got, want)
	}

	directProfile := filepath.Join(customRoot, "profiles", "direct")
	t.Setenv("HERMES_HOME", directProfile)
	if got, want := defaultHermesDatabase(home), filepath.Join(directProfile, "state.db"); got != want {
		t.Fatalf("profile HERMES_HOME path = %q, want %q", got, want)
	}

	t.Setenv("HERMES_HOME", "")
	base := filepath.Join(home, ".hermes")
	if err := os.MkdirAll(base, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(base, "active_profile"), []byte("coder\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	want := filepath.Join(base, "profiles", "coder", "state.db")
	if got := defaultHermesDatabase(home); got != want {
		t.Fatalf("profile path = %q, want %q", got, want)
	}
}

func TestDefaultOpenCodeDatabaseHonorsOverrides(t *testing.T) {
	home := t.TempDir()
	xdg := filepath.Join(home, "data")
	t.Setenv("XDG_DATA_HOME", xdg)
	t.Setenv("OPENCODE_DB", "custom.db")
	want := filepath.Join(xdg, "opencode", "custom.db")
	if got := defaultOpenCodeDatabase(home); got != want {
		t.Fatalf("relative override = %q, want %q", got, want)
	}

	absolute := filepath.Join(home, "absolute.db")
	t.Setenv("OPENCODE_DB", absolute)
	if got := defaultOpenCodeDatabase(home); got != absolute {
		t.Fatalf("absolute override = %q", got)
	}
}

func TestOpenCodeDataDirDefaultsToXDGConvention(t *testing.T) {
	home := t.TempDir()
	t.Setenv("XDG_DATA_HOME", "")
	if got, want := openCodeDataDir(home), filepath.Join(home, ".local", "share"); got != want {
		t.Fatalf("data dir = %q, want %q", got, want)
	}
}

func TestHermesBaseDirUsesWindowsLocalAppData(t *testing.T) {
	home := `C:\Users\test`
	if got, want := hermesBaseDir(home, "windows", `D:\Local`), filepath.Join(`D:\Local`, "hermes"); got != want {
		t.Fatalf("Hermes Windows base = %q, want %q", got, want)
	}
}

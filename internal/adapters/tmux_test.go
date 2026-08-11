package adapters

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestDetectAgentCommandRejectsOrdinaryShells(t *testing.T) {
	for _, command := range []string{"", "zsh", "bash", "fish", "node", "python"} {
		if agent, ok := detectAgentCommand(command); ok {
			t.Fatalf("ordinary command %q detected as %q", command, agent)
		}
	}
}

func TestDetectAgentCommandRecognizesKnownAgents(t *testing.T) {
	tests := map[string]string{
		"claude":       "claude",
		"copilot":      "copilot",
		"codex":        "codex",
		"hermes":       "hermes",
		"cursor-agent": "cursor-agent",
	}
	for command, want := range tests {
		got, ok := detectAgentCommand(command)
		if !ok || got != want {
			t.Fatalf("detectAgentCommand(%q) = %q, %v; want %q, true", command, got, ok, want)
		}
	}
}

func TestTmuxScannerIgnoresOrdinaryPanes(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("shell fixture is Unix-only")
	}
	binDir := t.TempDir()
	script := `#!/bin/sh
printf 'work\0370\0370\037pane-shell\037/repo/shell\037zsh\037shell\0371\n'
printf 'work\0370\0371\037pane-agent\037/repo/agent\037claude\037agent\0370\n'
`
	if err := os.WriteFile(filepath.Join(binDir, "tmux"), []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", binDir+string(os.PathListSeparator)+os.Getenv("PATH"))

	result := (TmuxScanner{}).Scan(context.Background())
	if result.Err != nil || !result.SnapshotComplete {
		t.Fatalf("scan failed: %+v", result)
	}
	if len(result.Runtimes) != 1 || result.Runtimes[0].NativeID != "pane-agent" {
		t.Fatalf("unexpected runtimes: %+v", result.Runtimes)
	}
	if len(result.Sessions) != 1 || result.Sessions[0].Agent != "claude" {
		t.Fatalf("unexpected sessions: %+v", result.Sessions)
	}
}

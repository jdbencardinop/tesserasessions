package adapters

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestObjectListReadsNestedResultAgents(t *testing.T) {
	var payload any
	data := []byte(`{
		"result": {
			"agents": [
				{
					"name": "auth-agent",
					"pane_id": "w1:p1",
					"kind": "claude",
					"agent_status": "blocked",
					"foreground_cwd": "/repo",
					"agent_session": {
						"source": "hook",
						"agent": "claude",
						"kind": "id",
						"value": "sess-1"
					}
				}
			]
		}
	}`)
	if err := json.Unmarshal(data, &payload); err != nil {
		t.Fatal(err)
	}
	items, recognized := herdrAgentItems(payload)
	if !recognized {
		t.Fatal("expected recognized Herdr response")
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items))
	}
	target, pane := herdrTarget(items[0])
	if target != "auth-agent" {
		t.Fatalf("target = %q, want auth-agent", target)
	}
	if pane != "w1:p1" {
		t.Fatalf("pane = %q, want w1:p1", pane)
	}
	if got := normalizeStatus(firstString(items[0], "agent_status")); got != "needs_attention" {
		t.Fatalf("status = %q", got)
	}
	ref, agent := agentSessionRef(items[0])
	if ref != "sess-1" || agent != "claude" {
		t.Fatalf("session ref = %q, agent = %q", ref, agent)
	}
}

func TestHerdrAgentItemsRecognizesExplicitEmptyList(t *testing.T) {
	items, recognized := herdrAgentItems(map[string]any{
		"result": map[string]any{"agents": []any{}},
	})
	if !recognized || len(items) != 0 {
		t.Fatalf("explicit empty list not recognized: recognized=%v items=%v", recognized, items)
	}
}

func TestHerdrAgentItemsRejectsUnknownShape(t *testing.T) {
	if items, recognized := herdrAgentItems(map[string]any{}); recognized || items != nil {
		t.Fatalf("unknown shape accepted: recognized=%v items=%v", recognized, items)
	}
	if items, recognized := herdrAgentItems(map[string]any{"id": "cli:agent:list"}); recognized || items != nil {
		t.Fatalf("envelope id accepted as an agent: recognized=%v items=%v", recognized, items)
	}
}

func TestHerdrAgentListUsesDefaultJSONOutput(t *testing.T) {
	args := herdrAgentListArgs()
	if len(args) != 2 || args[0] != "agent" || args[1] != "list" {
		t.Fatalf("unexpected Herdr args: %v", args)
	}
}

func TestHerdrScannerAcceptsCurrentCLIShape(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("shell fixture is Unix-only")
	}
	binDir := t.TempDir()
	script := `#!/bin/sh
if [ "$#" -ne 2 ] || [ "$1" != "agent" ] || [ "$2" != "list" ]; then
  exit 2
fi
printf '%s\n' '{"result":{"agents":[{"name":"auth-agent","pane_id":"w1:p1","kind":"claude","agent_status":"blocked","foreground_cwd":"/repo"}]}}'
`
	if err := os.WriteFile(filepath.Join(binDir, "herdr"), []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", binDir+string(os.PathListSeparator)+os.Getenv("PATH"))

	result := (HerdrScanner{}).Scan(context.Background())
	if result.Err != nil || !result.SnapshotComplete {
		t.Fatalf("scan failed: %+v", result)
	}
	if len(result.Runtimes) != 1 || result.Runtimes[0].Status != "needs_attention" {
		t.Fatalf("unexpected runtimes: %+v", result.Runtimes)
	}
}

func TestResumeCommandForKnownAgents(t *testing.T) {
	tests := map[string]string{
		"claude":  "claude --resume 'abc'",
		"copilot": "copilot --resume='abc'",
		"hermes":  "hermes --resume 'abc'",
	}
	for agent, want := range tests {
		if got := resumeCommand(agent, "abc"); got != want {
			t.Fatalf("resumeCommand(%q) = %q, want %q", agent, got, want)
		}
	}
}

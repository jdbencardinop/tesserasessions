package adapters

import (
	"encoding/json"
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
					"state": "blocked",
					"foreground_cwd": "/repo",
					"agent_session": {"id": "sess-1"}
				}
			]
		}
	}`)
	if err := json.Unmarshal(data, &payload); err != nil {
		t.Fatal(err)
	}
	items := objectList(payload)
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
	if got := normalizeStatus(firstString(items[0], "state")); got != "needs_attention" {
		t.Fatalf("status = %q", got)
	}
	if got := agentSessionRef(items[0]); got != "sess-1" {
		t.Fatalf("session ref = %q", got)
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

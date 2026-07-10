package adapters

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestClaudeScannerFindsJSONL(t *testing.T) {
	root := t.TempDir()
	project := filepath.Join(root, "-tmp-project")
	if err := os.MkdirAll(project, 0755); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(project, "session-1.jsonl")
	data := `{"timestamp":"2026-01-01T00:00:00Z","message":{"role":"user","content":"Build auth"}}` + "\n"
	if err := os.WriteFile(path, []byte(data), 0644); err != nil {
		t.Fatal(err)
	}
	res := ClaudeScanner{Root: root}.Scan(context.Background())
	if res.Err != nil {
		t.Fatal(res.Err)
	}
	if len(res.Sessions) != 1 {
		t.Fatalf("expected one session, got %d", len(res.Sessions))
	}
	if res.Sessions[0].Agent != "claude" {
		t.Fatalf("expected claude agent, got %q", res.Sessions[0].Agent)
	}
}

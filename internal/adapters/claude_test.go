package adapters

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/jdbencardinop/tesserasessions/internal/core"
)

func TestClaudeScannerUsesExactTranscriptMetadata(t *testing.T) {
	root := t.TempDir()
	project := filepath.Join(root, "-Users-test-my-project-with-hyphen")
	if err := os.MkdirAll(project, 0o755); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(project, "exact-id.jsonl")
	data := strings.Join([]string{
		`{"type":"user","sessionId":"exact-id","cwd":"/Users/test/my-project-with-hyphen","gitBranch":"feature/auth","timestamp":"2026-01-01T00:00:00Z"}`,
		`{"type":"ai-title","sessionId":"exact-id","aiTitle":"AI title"}`,
		`{"type":"custom-title","sessionId":"exact-id","customTitle":"User title"}`,
		`{"type":"assistant","sessionId":"exact-id","cwd":"/Users/test/my-project-with-hyphen","timestamp":"2026-01-01T01:00:00Z"}`,
		`{"type":"custom-title","sessionId":"sidechain-id","cwd":"/wrong","customTitle":"Wrong sidechain title","timestamp":"2026-01-02T01:00:00Z"}`,
		`{"partial":`,
	}, "\n")
	if err := os.WriteFile(path, []byte(data), 0o644); err != nil {
		t.Fatal(err)
	}

	result := (ClaudeScanner{Root: root}).Scan(context.Background())
	if result.Err != nil {
		t.Fatal(result.Err)
	}
	if len(result.Sessions) != 1 {
		t.Fatalf("expected one session, got %d", len(result.Sessions))
	}
	session := result.Sessions[0]
	if session.NativeID != "exact-id" {
		t.Fatalf("native id = %q", session.NativeID)
	}
	if session.ProjectPath != "/Users/test/my-project-with-hyphen" {
		t.Fatalf("project path = %q", session.ProjectPath)
	}
	if session.Title != "User title" {
		t.Fatalf("title = %q", session.Title)
	}
	if session.CreatedAt != time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC) {
		t.Fatalf("created = %s", session.CreatedAt)
	}
	if session.LastActivityAt != time.Date(2026, 1, 1, 1, 0, 0, 0, time.UTC) {
		t.Fatalf("updated = %s", session.LastActivityAt)
	}
	wantResume := "cd '/Users/test/my-project-with-hyphen' && CLAUDE_CONFIG_DIR=" +
		core.ShellQuote(filepath.Dir(root)) + " claude --resume 'exact-id'"
	if session.ResumeCommand != wantResume {
		t.Fatalf("resume = %q, want %q", session.ResumeCommand, wantResume)
	}
}

func TestClaudeScannerSkipsOversizedRowsAndContinues(t *testing.T) {
	root := t.TempDir()
	project := filepath.Join(root, "-repo")
	if err := os.MkdirAll(project, 0o755); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(project, "session.jsonl")
	oversized := `{"type":"assistant","sessionId":"session","payload":"` +
		strings.Repeat("x", 2*1024*1024+1024) + `"}`
	data := oversized + "\n" +
		`{"type":"custom-title","sessionId":"session","cwd":"/repo","customTitle":"After large row","timestamp":"2026-01-01T00:00:00Z"}` + "\n"
	if err := os.WriteFile(path, []byte(data), 0o644); err != nil {
		t.Fatal(err)
	}
	result := (ClaudeScanner{Root: root}).Scan(context.Background())
	if result.Err != nil || len(result.Sessions) != 1 {
		t.Fatalf("unexpected result: %+v", result)
	}
	if result.Sessions[0].Title != "After large row" {
		t.Fatalf("title after oversized row = %q", result.Sessions[0].Title)
	}
}

func TestClaudeScannerFallsBackToEncodedDirectory(t *testing.T) {
	root := t.TempDir()
	project := filepath.Join(root, "-tmp-project")
	if err := os.MkdirAll(project, 0o755); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(project, "session-1.jsonl")
	if err := os.WriteFile(path, []byte(`{"timestamp":"2026-01-01T00:00:00Z"}`+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	result := (ClaudeScanner{Root: root}).Scan(context.Background())
	if result.Err != nil || len(result.Sessions) != 1 {
		t.Fatalf("unexpected result: %+v", result)
	}
	if result.Sessions[0].ProjectPath != "/tmp/project" {
		t.Fatalf("fallback path = %q", result.Sessions[0].ProjectPath)
	}
}

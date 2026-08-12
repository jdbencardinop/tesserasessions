package adapters

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/jdbencardinop/tesserasessions/internal/core"
)

func TestCopilotScannerReadsWorkspaceMetadata(t *testing.T) {
	root := t.TempDir()
	writeCopilotWorkspace(t, root, "session-named", `
id: exact-id
cwd: /repo/worktree
git_root: /repo
repository: owner/repo
branch: feature/auth
name: Refactor auth
created_at: 2026-01-01T00:00:00Z
updated_at: 2026-01-01T02:00:00Z
`)
	writeCopilotWorkspace(t, root, "session-unnamed", `
id: fallback-id
cwd: /repo/other
repository: owner/other
created_at: 2026-01-02T00:00:00Z
updated_at: 2026-01-02T01:00:00Z
`)
	if err := os.Mkdir(filepath.Join(root, "not-a-session"), 0o755); err != nil {
		t.Fatal(err)
	}

	result := (CopilotScanner{Root: root}).Scan(context.Background())
	if result.Err != nil {
		t.Fatal(result.Err)
	}
	if len(result.Sessions) != 2 {
		t.Fatalf("sessions = %d", len(result.Sessions))
	}
	named := findNativeSession(t, result.Sessions, "exact-id")
	if named.ProjectPath != "/repo/worktree" || named.Title != "Refactor auth" {
		t.Fatalf("unexpected named session: %+v", named)
	}
	if named.CreatedAt != time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC) ||
		named.LastActivityAt != time.Date(2026, 1, 1, 2, 0, 0, 0, time.UTC) {
		t.Fatalf("unexpected timestamps: %+v", named)
	}
	wantResume := "cd '/repo/worktree' && COPILOT_HOME=" +
		core.ShellQuote(filepath.Dir(root)) + " copilot --resume='exact-id'"
	if named.ResumeCommand != wantResume {
		t.Fatalf("resume = %q, want %q", named.ResumeCommand, wantResume)
	}
	unnamed := findNativeSession(t, result.Sessions, "fallback-id")
	if unnamed.Title != "Copilot: owner/other" {
		t.Fatalf("fallback title = %q", unnamed.Title)
	}
}

func TestCopilotScannerRejectsMalformedWorkspace(t *testing.T) {
	root := t.TempDir()
	writeCopilotWorkspace(t, root, "broken", "id: [\n")
	result := (CopilotScanner{Root: root}).Scan(context.Background())
	if result.Err == nil {
		t.Fatal("expected malformed workspace error")
	}
}

func writeCopilotWorkspace(t *testing.T, root, id, content string) {
	t.Helper()
	dir := filepath.Join(root, id)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "workspace.yaml"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

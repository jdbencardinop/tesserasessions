package cli

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/jdbencardinop/tesserasessions/internal/adapters"
	"github.com/jdbencardinop/tesserasessions/internal/core"
	"github.com/jdbencardinop/tesserasessions/internal/store"
)

func TestPersistScanRuntimesOnlyPrunesCompleteSnapshots(t *testing.T) {
	db, err := store.Open(filepath.Join(t.TempDir(), "sessions.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close() //nolint:errcheck

	ctx := context.Background()
	runtime := core.RuntimeInstance{
		ID:        "tmux-1",
		SessionID: "session-1",
		Backend:   "tmux",
		NativeID:  "%1",
		Status:    core.StatusIdle,
	}
	if err := db.UpsertRuntime(ctx, runtime); err != nil {
		t.Fatal(err)
	}

	if err := persistScanRuntimes(ctx, db, core.ScanResult{
		Source:  "tmux",
		Skipped: true,
		Message: "tmux unavailable",
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := db.RuntimeForSession(ctx, "session-1", "tmux"); err != nil {
		t.Fatalf("incomplete scan pruned prior row: %v", err)
	}

	if err := persistScanRuntimes(ctx, db, core.ScanResult{
		Source:           "tmux",
		SnapshotComplete: true,
		Err:              errors.New("provider failed after producing partial data"),
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := db.RuntimeForSession(ctx, "session-1", "tmux"); err != nil {
		t.Fatalf("failed complete scan pruned prior row: %v", err)
	}

	if err := persistScanRuntimes(ctx, db, core.ScanResult{
		Source:           "tmux",
		SnapshotComplete: true,
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := db.RuntimeForSession(ctx, "session-1", "tmux"); !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("complete empty scan should prune row, got %v", err)
	}
}

func TestPersistScanSessionsOnlyPrunesCompleteSnapshots(t *testing.T) {
	db, err := store.Open(filepath.Join(t.TempDir(), "sessions.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close() //nolint:errcheck
	ctx := context.Background()

	session := core.NewSession("opencode", "session-1")
	session.Title = "Session"
	if err := db.UpsertSession(ctx, session); err != nil {
		t.Fatal(err)
	}
	if err := persistScanSessions(ctx, db, core.ScanResult{
		Source:  "opencode",
		Skipped: true,
		Message: "database unavailable",
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := db.GetSession(ctx, session.ID); err != nil {
		t.Fatalf("incomplete scan pruned prior session: %v", err)
	}

	if err := persistScanSessions(ctx, db, core.ScanResult{
		Source:                  "opencode",
		SessionSnapshotComplete: true,
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := db.GetSession(ctx, session.ID); !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("complete empty scan should prune session, got %v", err)
	}
}

func TestCodexAuthoritativeScanReconcilesRemovedRollouts(t *testing.T) {
	home := t.TempDir()
	sessions := filepath.Join(home, "sessions")
	if err := os.MkdirAll(sessions, 0o755); err != nil {
		t.Fatal(err)
	}
	rollout := filepath.Join(sessions, "rollout-2026-01-01T00-00-00-thread.jsonl")
	line := `{"timestamp":"2026-01-01T00:00:00Z","type":"session_meta","payload":{"id":"thread","cwd":"/repo","timestamp":"2026-01-01T00:00:00Z","source":"cli","history_mode":"legacy"}}`
	if err := os.WriteFile(rollout, []byte(line+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	db, err := store.Open(filepath.Join(t.TempDir(), "sessions.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close() //nolint:errcheck
	ctx := context.Background()
	scanner := adapters.CodexScanner{Home: home}
	first := scanner.Scan(ctx)
	if first.Err != nil || len(first.Sessions) != 1 {
		t.Fatalf("initial scan: %+v", first)
	}
	if err := persistScanSessions(ctx, db, first); err != nil {
		t.Fatal(err)
	}
	sessionID := first.Sessions[0].ID

	if err := os.Remove(rollout); err != nil {
		t.Fatal(err)
	}
	malformed := filepath.Join(sessions, "rollout-in-progress.jsonl")
	if err := os.WriteFile(malformed, []byte("{"), 0o644); err != nil {
		t.Fatal(err)
	}
	newRollout := filepath.Join(sessions, "rollout-new.jsonl")
	newLine := `{"timestamp":"2026-01-02T00:00:00Z","type":"session_meta","payload":{"id":"new-thread","cwd":"/repo/new","timestamp":"2026-01-02T00:00:00Z","source":"cli","history_mode":"legacy"}}`
	if err := os.WriteFile(newRollout, []byte(newLine+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	incomplete := scanner.Scan(ctx)
	if incomplete.Err != nil || incomplete.SessionSnapshotComplete || incomplete.Message == "" || len(incomplete.Sessions) != 1 {
		t.Fatalf("incomplete scan: %+v", incomplete)
	}
	if err := persistScanSessions(ctx, db, incomplete); err != nil {
		t.Fatal(err)
	}
	newSessionID := incomplete.Sessions[0].ID
	if _, err := db.GetSession(ctx, newSessionID); err != nil {
		t.Fatalf("incomplete Codex scan did not refresh valid row: %v", err)
	}
	if _, err := db.GetSession(ctx, sessionID); err != nil {
		t.Fatalf("incomplete Codex scan pruned prior row: %v", err)
	}

	if err := os.Remove(malformed); err != nil {
		t.Fatal(err)
	}
	complete := scanner.Scan(ctx)
	if complete.Err != nil || !complete.SessionSnapshotComplete || len(complete.Sessions) != 1 {
		t.Fatalf("empty authoritative scan: %+v", complete)
	}
	if err := persistScanSessions(ctx, db, complete); err != nil {
		t.Fatal(err)
	}
	if _, err := db.GetSession(ctx, sessionID); !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("removed Codex rollout remained in inventory: %v", err)
	}
	if _, err := db.GetSession(ctx, newSessionID); err != nil {
		t.Fatalf("valid Codex rollout was lost after complete scan: %v", err)
	}
}

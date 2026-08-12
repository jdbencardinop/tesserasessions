package cli

import (
	"context"
	"database/sql"
	"errors"
	"path/filepath"
	"testing"

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

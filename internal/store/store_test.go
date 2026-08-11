package store

import (
	"context"
	"database/sql"
	"errors"
	"path/filepath"
	"testing"
	"time"

	"github.com/jdbencardinop/tesserasessions/internal/core"
)

func TestUpsertAndGetSession(t *testing.T) {
	db, err := Open(filepath.Join(t.TempDir(), "sessions.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close() //nolint:errcheck
	sess := core.NewSession("test", "native")
	sess.ProjectPath = t.TempDir()
	sess.Agent = "agent"
	sess.Title = "Test session"
	sess.LastActivityAt = time.Now().UTC()
	if err := db.UpsertSession(context.Background(), sess); err != nil {
		t.Fatal(err)
	}
	got, err := db.GetSession(context.Background(), sess.ID[:8])
	if err != nil {
		t.Fatal(err)
	}
	if got.ID != sess.ID || got.Title != sess.Title {
		t.Fatalf("unexpected session: %#v", got)
	}
}

func TestReplaceRuntimesIsAtomicPerBackend(t *testing.T) {
	db, err := Open(filepath.Join(t.TempDir(), "sessions.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close() //nolint:errcheck
	ctx := context.Background()

	for _, runtime := range []core.RuntimeInstance{
		{ID: "tmux-1", SessionID: "session-1", Backend: "tmux", NativeID: "%1"},
		{ID: "tmux-2", SessionID: "session-2", Backend: "tmux", NativeID: "%2"},
		{ID: "herdr-1", SessionID: "session-3", Backend: "herdr", NativeID: "w1:p1"},
	} {
		if err := db.UpsertRuntime(ctx, runtime); err != nil {
			t.Fatal(err)
		}
	}

	if err := db.ReplaceRuntimes(ctx, "tmux", []core.RuntimeInstance{{
		ID:        "tmux-2",
		SessionID: "session-2",
		Backend:   "tmux",
		NativeID:  "%2",
		Status:    core.StatusWorking,
	}}); err != nil {
		t.Fatal(err)
	}
	if _, err := db.RuntimeForSession(ctx, "session-1", "tmux"); !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("removed tmux row still present: %v", err)
	}
	got, err := db.RuntimeForSession(ctx, "session-2", "tmux")
	if err != nil || got.Status != core.StatusWorking {
		t.Fatalf("replacement row missing: %+v, %v", got, err)
	}
	if _, err := db.RuntimeForSession(ctx, "session-3", "herdr"); err != nil {
		t.Fatalf("other backend was pruned: %v", err)
	}

	if err := db.ReplaceRuntimes(ctx, "tmux", nil); err != nil {
		t.Fatal(err)
	}
	if _, err := db.RuntimeForSession(ctx, "session-2", "tmux"); !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("empty complete snapshot did not clear backend: %v", err)
	}
}

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
	session := core.NewSession("test", "native")
	session.ProjectPath = t.TempDir()
	session.Agent = "agent"
	session.Title = "Test session"
	session.LastActivityAt = time.Now().UTC()
	if err := db.UpsertSession(context.Background(), session); err != nil {
		t.Fatal(err)
	}
	got, err := db.GetSession(context.Background(), session.ID[:8])
	if err != nil {
		t.Fatal(err)
	}
	if got.ID != session.ID || got.Title != session.Title {
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

func TestReplaceSessionsPreservesManualMetadataAndPrunes(t *testing.T) {
	db, err := Open(filepath.Join(t.TempDir(), "sessions.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close() //nolint:errcheck
	ctx := context.Background()

	keep := core.NewSession("hermes", "keep")
	keep.Title = "Scanner title"
	keep.ProjectPath = "/repo"
	keep.CreatedAt = time.Unix(100, 0).UTC()
	remove := core.NewSession("hermes", "remove")
	remove.Title = "Remove me"
	if err := db.ReplaceSessions(ctx, "hermes", []core.Session{keep, remove}); err != nil {
		t.Fatal(err)
	}
	if err := db.UpdateTitle(ctx, keep.ID, "Manual title"); err != nil {
		t.Fatal(err)
	}
	if err := db.UpdatePinned(ctx, keep.ID, true); err != nil {
		t.Fatal(err)
	}
	if err := db.UpdateTags(ctx, keep.ID, "review,mvp"); err != nil {
		t.Fatal(err)
	}

	keep.Title = "New scanner title"
	keep.ProjectPath = "/repo/new"
	keep.CreatedAt = time.Unix(200, 0).UTC()
	if err := db.ReplaceSessions(ctx, "hermes", []core.Session{keep}); err != nil {
		t.Fatal(err)
	}
	got, err := db.GetSession(ctx, keep.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.Title != "Manual title" ||
		!got.Pinned ||
		got.Tags != "review,mvp" ||
		got.ProjectPath != "/repo/new" ||
		got.CreatedAt.Unix() != 200 {
		t.Fatalf("manual metadata or scanner update lost: %+v", got)
	}
	if _, err := db.GetSession(ctx, remove.ID); !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("removed source session still present: %v", err)
	}
}

func TestUpsertSessionPreservesLiveCreationTime(t *testing.T) {
	db, err := Open(filepath.Join(t.TempDir(), "sessions.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close() //nolint:errcheck
	ctx := context.Background()

	session := core.NewSession("tmux", "pane")
	session.Title = "Live"
	session.CreatedAt = time.Unix(100, 0).UTC()
	if err := db.UpsertSession(ctx, session); err != nil {
		t.Fatal(err)
	}
	session.CreatedAt = time.Unix(200, 0).UTC()
	if err := db.UpsertSession(ctx, session); err != nil {
		t.Fatal(err)
	}
	got, err := db.GetSession(ctx, session.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.CreatedAt.Unix() != 100 {
		t.Fatalf("live upsert reset creation time: %s", got.CreatedAt)
	}
}

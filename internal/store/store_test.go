package store

import (
	"context"
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

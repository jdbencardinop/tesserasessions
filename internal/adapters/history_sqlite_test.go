package adapters

import (
	"bytes"
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"testing"

	"github.com/jdbencardinop/tesserasessions/internal/core"
)

func TestHermesHistoryScannerReadsCurrentMetadataOnly(t *testing.T) {
	path := filepath.Join(t.TempDir(), "state.db")
	db := createSQLiteFixture(t, path, `
		CREATE TABLE sessions (
			id TEXT PRIMARY KEY,
			source TEXT NOT NULL,
			title TEXT,
			cwd TEXT,
			git_repo_root TEXT,
			started_at REAL NOT NULL,
			ended_at REAL,
			last_activity_at REAL,
			parent_session_id TEXT,
			archived INTEGER DEFAULT 0,
			end_reason TEXT,
			model_config TEXT
		);
		CREATE TABLE messages (
			id INTEGER PRIMARY KEY,
			session_id TEXT,
			content TEXT,
			timestamp REAL,
			active INTEGER DEFAULT 1
		);
	`)
	mustExec(t, db, `INSERT INTO sessions (
			id, source, title, cwd, git_repo_root, started_at, ended_at,
			last_activity_at, parent_session_id, archived, end_reason, model_config
		) VALUES
		('root', 'cli', 'Fix build', '/repo/worktree', '/repo', 1000.25, 1100, 1100, NULL, 0, 'compression', NULL),
		('child', 'cli', 'Fix build #2', '/repo/new-worktree', '/repo', 1201, NULL, 1300.5, 'root', 0, NULL, NULL),
		('delegate', 'cli', 'Wrong delegate', '/wrong/delegate', '/wrong', 1202, NULL, 9999, 'root', 0, NULL, '{"_delegate_from":"root"}'),
		('branch', 'cli', 'Wrong branch', '/wrong/branch', '/wrong', 1203, NULL, 9998, 'root', 0, NULL, '{"_branched_from":"root"}'),
		('tool', 'tool', 'Wrong tool', '/wrong/tool', '/wrong', 1204, NULL, 9997, 'root', 0, NULL, NULL),
		('orphan-delegate', 'cli', 'Orphan delegate', '/wrong/orphan', '/wrong', 1205, NULL, 9996, NULL, 0, NULL, '{"_delegate_from":"deleted"}'),
		('orphan-branch', 'cli', 'Orphan branch', '/wrong/orphan', '/wrong', 1206, NULL, 9995, NULL, 0, NULL, '{"_branched_from":"deleted"}'),
		('archived', 'cli', 'Old', '/repo', '/repo', 900, 950, 950, NULL, 1, NULL, NULL),
		('gateway', 'telegram', 'Chat', NULL, NULL, 1000, NULL, NULL, NULL, 0, NULL, NULL)`)
	mustExec(t, db, `INSERT INTO messages VALUES
		(1, 'root', 'private transcript content', 1150, 1),
		(2, 'child', 'latest private transcript content', 1400.5, 1)`)
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}
	before := readFixtureBytes(t, path)

	result := (HermesHistoryScanner{Database: path}).Scan(context.Background())
	if result.Err != nil {
		t.Fatal(result.Err)
	}
	if len(result.Sessions) != 1 {
		t.Fatalf("sessions = %d, want 1", len(result.Sessions))
	}
	session := result.Sessions[0]
	if session.NativeID != "root" || session.Title != "Fix build #2" || session.ProjectPath != "/repo/new-worktree" {
		t.Fatalf("unexpected session: %+v", session)
	}
	if session.CreatedAt.Unix() != 1000 || session.LastActivityAt.Unix() != 1400 {
		t.Fatalf("unexpected times: created=%s updated=%s", session.CreatedAt, session.LastActivityAt)
	}
	wantResume := "cd '/repo/new-worktree' && HERMES_HOME=" +
		core.ShellQuote(filepath.Dir(path)) +
		" HERMES_S6_SUPERVISED_CHILD='1' hermes --resume 'root'"
	if session.ResumeCommand != wantResume {
		t.Fatalf("resume = %q", session.ResumeCommand)
	}
	if !result.SessionSnapshotComplete {
		t.Fatal("successful Hermes scan must be authoritative")
	}
	if !bytes.Equal(before, readFixtureBytes(t, path)) {
		t.Fatal("Hermes scanner modified the database")
	}
}

func TestHermesResumeCommandKeepsProfileScopedHome(t *testing.T) {
	database := filepath.Join("/hermes", "profiles", "coder", "state.db")
	want := "HERMES_HOME='/hermes/profiles/coder' hermes --resume 'session'"
	if got := hermesResumeCommand(database, "session"); got != want {
		t.Fatalf("resume = %q, want %q", got, want)
	}
}

func TestOpenCodeScannerReadsCurrentMetadataOnly(t *testing.T) {
	path := filepath.Join(t.TempDir(), "opencode.db")
	db := createSQLiteFixture(t, path, `
		CREATE TABLE project (
			id TEXT PRIMARY KEY,
			worktree TEXT NOT NULL,
			name TEXT
		);
		CREATE TABLE "session" (
			id TEXT PRIMARY KEY,
			slug TEXT,
			title TEXT,
			project_id TEXT NOT NULL,
			parent_id TEXT,
			directory TEXT,
			path TEXT,
			time_created INTEGER NOT NULL,
			time_updated INTEGER NOT NULL,
			time_archived INTEGER
		);
	`)
	mustExec(t, db, `INSERT INTO project VALUES ('project', '/repo', 'app')`)
	mustExec(t, db, `INSERT INTO "session" VALUES
		('root', 'happy-dolphin', 'Refactor auth', 'project', NULL, '/repo/packages/api', 'packages/api', 1000000, 2000000, NULL),
		('child', 'child', 'Child', 'project', 'root', '/repo', NULL, 1000000, 2000000, NULL),
		('archived', 'old', 'Old', 'project', NULL, '/repo', NULL, 1000000, 2000000, 3000000),
		('fallback', 'quiet-panda', '', 'project', NULL, '', 'packages/web', 3000000, 4000000, NULL)`)
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}
	before := readFixtureBytes(t, path)

	result := (OpenCodeScanner{Database: path}).Scan(context.Background())
	if result.Err != nil {
		t.Fatal(result.Err)
	}
	if len(result.Sessions) != 2 {
		t.Fatalf("sessions = %d, want 2", len(result.Sessions))
	}
	root := findNativeSession(t, result.Sessions, "root")
	if root.ProjectPath != "/repo/packages/api" || root.Title != "Refactor auth" {
		t.Fatalf("unexpected root: %+v", root)
	}
	wantResume := "cd '/repo/packages/api' && OPENCODE_DB=" +
		core.ShellQuote(path) + " opencode --session 'root'"
	if root.ResumeCommand != wantResume {
		t.Fatalf("resume = %q", root.ResumeCommand)
	}
	fallback := findNativeSession(t, result.Sessions, "fallback")
	if fallback.ProjectPath != "/repo/packages/web" || fallback.Title != "quiet-panda" {
		t.Fatalf("unexpected fallback: %+v", fallback)
	}
	if !result.SessionSnapshotComplete {
		t.Fatal("successful OpenCode scan must be authoritative")
	}
	if !bytes.Equal(before, readFixtureBytes(t, path)) {
		t.Fatal("OpenCode scanner modified the database")
	}
}

func TestHistorySQLiteScannersSkipMissingAndRejectMalformed(t *testing.T) {
	missing := filepath.Join(t.TempDir(), "missing.db")
	for _, result := range []core.ScanResult{
		(HermesHistoryScanner{Database: missing}).Scan(context.Background()),
		(OpenCodeScanner{Database: missing}).Scan(context.Background()),
	} {
		if !result.Skipped || result.Err != nil {
			t.Fatalf("missing database result: %+v", result)
		}
	}

	broken := filepath.Join(t.TempDir(), "broken.db")
	if err := os.WriteFile(broken, []byte("not sqlite"), 0o644); err != nil {
		t.Fatal(err)
	}
	before := readFixtureBytes(t, broken)
	for _, result := range []core.ScanResult{
		(HermesHistoryScanner{Database: broken}).Scan(context.Background()),
		(OpenCodeScanner{Database: broken}).Scan(context.Background()),
	} {
		if result.Err == nil {
			t.Fatalf("malformed database accepted: %+v", result)
		}
	}
	if !bytes.Equal(before, readFixtureBytes(t, broken)) {
		t.Fatal("malformed database was modified")
	}
}

func TestOpenCodeScannerRejectsOrphanedSessions(t *testing.T) {
	path := filepath.Join(t.TempDir(), "opencode.db")
	db := createSQLiteFixture(t, path, `
		CREATE TABLE project (
			id TEXT PRIMARY KEY,
			worktree TEXT NOT NULL,
			name TEXT
		);
		CREATE TABLE "session" (
			id TEXT PRIMARY KEY,
			slug TEXT,
			title TEXT,
			project_id TEXT NOT NULL,
			parent_id TEXT,
			directory TEXT,
			path TEXT,
			time_created INTEGER NOT NULL,
			time_updated INTEGER NOT NULL,
			time_archived INTEGER
		);
	`)
	mustExec(t, db, `INSERT INTO "session" VALUES
		('orphan', 'orphan', 'Orphan', 'missing', NULL, '/repo', NULL, 1000, 2000, NULL)`)
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}
	result := (OpenCodeScanner{Database: path}).Scan(context.Background())
	if result.Err == nil {
		t.Fatal("expected orphan relationship error")
	}
}

func TestReadOnlySQLiteDoesNotMutateLiveWALFiles(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "live.db")
	writer, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatal(err)
	}
	defer writer.Close() //nolint:errcheck
	mustExec(t, writer, `PRAGMA journal_mode=WAL`)
	mustExec(t, writer, `PRAGMA wal_autocheckpoint=0`)
	mustExec(t, writer, `CREATE TABLE sessions (id TEXT PRIMARY KEY)`)
	mustExec(t, writer, `INSERT INTO sessions VALUES ('from-wal')`)

	before := snapshotDirectory(t, dir)
	scanPath := path
	symlink := filepath.Join(t.TempDir(), "linked.db")
	if err := os.Symlink(path, symlink); err == nil {
		scanPath = symlink
	}
	reader, err := openReadOnlySQLite(context.Background(), scanPath)
	if err != nil {
		t.Fatal(err)
	}
	var id string
	if err := reader.QueryRow(`SELECT id FROM sessions`).Scan(&id); err != nil {
		t.Fatal(err)
	}
	if id != "from-wal" {
		t.Fatalf("read id = %q", id)
	}
	if _, err := reader.Exec(`INSERT INTO sessions VALUES ('write')`); err == nil {
		t.Fatal("private snapshot connection accepted a write")
	}
	if err := reader.Close(); err != nil {
		t.Fatal(err)
	}
	after := snapshotDirectory(t, dir)
	if len(before) != len(after) {
		t.Fatalf("directory entries changed: before=%v after=%v", mapKeys(before), mapKeys(after))
	}
	for name, want := range before {
		if got, ok := after[name]; !ok || !bytes.Equal(got, want) {
			t.Fatalf("SQLite source sidecar %q changed", name)
		}
	}
}

func TestReadOnlySQLiteRecoversPrivateRollbackJournal(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "rollback.db")
	writer, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatal(err)
	}
	defer writer.Close() //nolint:errcheck
	mustExec(t, writer, `PRAGMA journal_mode=DELETE`)
	mustExec(t, writer, `CREATE TABLE items (value TEXT)`)
	mustExec(t, writer, `INSERT INTO items VALUES ('committed')`)

	tx, err := writer.Begin()
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback() //nolint:errcheck
	if _, err := tx.Exec(`UPDATE items SET value = 'uncommitted'`); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(path + "-journal"); err != nil {
		t.Fatalf("rollback journal missing: %v", err)
	}

	before := snapshotDirectory(t, dir)
	reader, err := openReadOnlySQLite(context.Background(), path)
	if err != nil {
		t.Fatal(err)
	}
	var value string
	if err := reader.QueryRow(`SELECT value FROM items`).Scan(&value); err != nil {
		t.Fatal(err)
	}
	if value != "committed" {
		t.Fatalf("snapshot exposed %q, want committed value", value)
	}
	if err := reader.Close(); err != nil {
		t.Fatal(err)
	}
	after := snapshotDirectory(t, dir)
	for name, want := range before {
		if got, ok := after[name]; !ok || !bytes.Equal(got, want) {
			t.Fatalf("rollback source file %q changed", name)
		}
	}
}

func TestSQLiteFileURIHandlesWindowsPaths(t *testing.T) {
	if got, want := sqliteFileURIForOS(`C:\Temp\source.db`, "windows"), "file:///C:/Temp/source.db"; got != want {
		t.Fatalf("drive URI = %q, want %q", got, want)
	}
	if got, want := sqliteFileURIForOS(`\\server\share\source.db`, "windows"), "file://server/share/source.db"; got != want {
		t.Fatalf("UNC URI = %q, want %q", got, want)
	}
}

func createSQLiteFixture(t *testing.T, path, schema string) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatal(err)
	}
	mustExec(t, db, schema)
	return db
}

func mustExec(t *testing.T, db *sql.DB, query string) {
	t.Helper()
	if _, err := db.Exec(query); err != nil {
		t.Fatal(err)
	}
}

func readFixtureBytes(t *testing.T, path string) []byte {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return data
}

func snapshotDirectory(t *testing.T, dir string) map[string][]byte {
	t.Helper()
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	result := make(map[string][]byte, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		result[entry.Name()] = readFixtureBytes(t, filepath.Join(dir, entry.Name()))
	}
	return result
}

func mapKeys(values map[string][]byte) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	return keys
}

func findNativeSession(t *testing.T, sessions []core.Session, nativeID string) core.Session {
	t.Helper()
	for _, session := range sessions {
		if session.NativeID == nativeID {
			return session
		}
	}
	t.Fatalf("session %q not found", nativeID)
	return core.Session{}
}

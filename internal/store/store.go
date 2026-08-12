package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jdbencardinop/tesserasessions/internal/core"
	_ "modernc.org/sqlite"
)

type Store struct {
	db *sql.DB
}

type Filter struct {
	Source     string
	Status     string
	Query      string
	Tag        string
	Sort       string
	Limit      int
	PinnedOnly bool
}

func Open(path string) (*Store, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)
	s := &Store{db: db}
	if err := s.Migrate(context.Background()); err != nil {
		_ = db.Close()
		return nil, err
	}
	return s, nil
}

func (s *Store) Close() error {
	return s.db.Close()
}

func (s *Store) Migrate(ctx context.Context) error {
	stmts := []string{
		`PRAGMA foreign_keys = ON`,
		`CREATE TABLE IF NOT EXISTS sources (
			id TEXT PRIMARY KEY,
			kind TEXT NOT NULL,
			path TEXT,
			enabled INTEGER NOT NULL DEFAULT 1,
			last_scan_at TEXT,
			last_error TEXT
		)`,
		`CREATE TABLE IF NOT EXISTS projects (
			id TEXT PRIMARY KEY,
			path TEXT NOT NULL UNIQUE,
			name TEXT NOT NULL,
			updated_at TEXT NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS sessions (
			id TEXT PRIMARY KEY,
			source TEXT NOT NULL,
			native_id TEXT NOT NULL,
			project_id TEXT,
			project_path TEXT,
			agent TEXT,
			title TEXT,
			goal_summary TEXT,
			status TEXT NOT NULL,
			last_activity_at TEXT,
			created_at TEXT NOT NULL,
			updated_at TEXT NOT NULL,
			resume_command TEXT,
			attach_command TEXT,
			raw_path TEXT,
			UNIQUE(source, native_id)
		)`,
		`CREATE TABLE IF NOT EXISTS runtime_instances (
			id TEXT PRIMARY KEY,
			session_id TEXT,
			backend TEXT NOT NULL,
			native_id TEXT NOT NULL,
			surface TEXT,
			project_path TEXT,
			command TEXT,
			status TEXT NOT NULL,
			attach_command TEXT,
			updated_at TEXT NOT NULL,
			UNIQUE(backend, native_id)
		)`,
		`CREATE TABLE IF NOT EXISTS summaries (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			session_id TEXT NOT NULL,
			provider TEXT NOT NULL,
			title TEXT,
			goal_summary TEXT,
			confidence REAL NOT NULL DEFAULT 0,
			created_at TEXT NOT NULL
		)`,
		`CREATE INDEX IF NOT EXISTS idx_sessions_source ON sessions(source)`,
		`CREATE INDEX IF NOT EXISTS idx_sessions_status ON sessions(status)`,
		`CREATE INDEX IF NOT EXISTS idx_sessions_last_activity ON sessions(last_activity_at DESC)`,
	}
	for _, stmt := range stmts {
		if _, err := s.db.ExecContext(ctx, stmt); err != nil {
			return err
		}
	}
	for _, col := range []struct {
		name string
		typ  string
	}{
		{"title_source", "TEXT NOT NULL DEFAULT 'scanner'"},
		{"pinned", "INTEGER NOT NULL DEFAULT 0"},
		{"tags", "TEXT NOT NULL DEFAULT ''"},
	} {
		ok, err := s.columnExists(ctx, "sessions", col.name)
		if err != nil {
			return err
		}
		if !ok {
			if _, err := s.db.ExecContext(ctx, `ALTER TABLE sessions ADD COLUMN `+col.name+` `+col.typ); err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *Store) UpsertSource(ctx context.Context, id, kind, path, lastError string) error {
	now := time.Now().UTC().Format(time.RFC3339Nano)
	_, err := s.db.ExecContext(ctx, `INSERT INTO sources (id, kind, path, enabled, last_scan_at, last_error)
		VALUES (?, ?, ?, 1, ?, ?)
		ON CONFLICT(id) DO UPDATE SET kind = excluded.kind, path = excluded.path,
			last_scan_at = excluded.last_scan_at, last_error = excluded.last_error`,
		id, kind, path, now, lastError)
	return err
}

func (s *Store) UpsertSession(ctx context.Context, sess core.Session) error {
	if strings.TrimSpace(sess.Source) == "" || strings.TrimSpace(sess.NativeID) == "" {
		return fmt.Errorf("session source and native id are required")
	}
	return upsertSession(ctx, s.db, sess, false)
}

// ReplaceSessions atomically replaces one source's authoritative historical
// snapshot. Existing manual titles, pins, and tags survive rows that remain.
func (s *Store) ReplaceSessions(ctx context.Context, source string, sessions []core.Session) error {
	source = strings.TrimSpace(source)
	if source == "" {
		return fmt.Errorf("session source is required")
	}
	normalized := make([]core.Session, len(sessions))
	seen := make(map[string]struct{}, len(sessions))
	for index, session := range sessions {
		if session.Source == "" {
			session.Source = source
		}
		if session.Source != source {
			return fmt.Errorf("session source %q does not match snapshot source %q", session.Source, source)
		}
		session.NativeID = strings.TrimSpace(session.NativeID)
		if session.NativeID == "" {
			return fmt.Errorf("session native id is required")
		}
		if _, exists := seen[session.NativeID]; exists {
			return fmt.Errorf("duplicate native session id %q for source %q", session.NativeID, source)
		}
		seen[session.NativeID] = struct{}{}
		normalized[index] = session
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback() //nolint:errcheck

	for _, session := range normalized {
		if err := upsertSession(ctx, tx, session, true); err != nil {
			return err
		}
	}

	args := []any{source}
	query := `DELETE FROM sessions WHERE source = ?`
	if len(normalized) > 0 {
		placeholders := make([]string, 0, len(normalized))
		for _, session := range normalized {
			placeholders = append(placeholders, "?")
			args = append(args, session.NativeID)
		}
		query += ` AND native_id NOT IN (` + strings.Join(placeholders, ",") + `)`
	}
	if _, err := tx.ExecContext(ctx, query, args...); err != nil {
		return err
	}
	return tx.Commit()
}

func upsertSession(ctx context.Context, executor dbExecer, sess core.Session, replaceCreatedAt bool) error {
	now := time.Now().UTC()
	sess.UpdatedAt = now
	if sess.CreatedAt.IsZero() {
		sess.CreatedAt = now
	}
	if sess.Status == "" {
		sess.Status = core.StatusUnknown
	}
	projectID := ""
	if sess.ProjectPath != "" {
		projectID = core.ProjectID(sess.ProjectPath)
		if err := upsertProject(ctx, executor, projectID, sess.ProjectPath); err != nil {
			return err
		}
	}
	titleSource := sess.TitleSource
	if titleSource == "" {
		titleSource = "scanner"
	}
	_, err := executor.ExecContext(ctx, `INSERT INTO sessions
		(id, source, native_id, project_id, project_path, agent, title, title_source, goal_summary, status,
		 pinned, tags, last_activity_at, created_at, updated_at, resume_command, attach_command, raw_path)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(source, native_id) DO UPDATE SET
			project_id = excluded.project_id,
			project_path = excluded.project_path,
			agent = excluded.agent,
			title = CASE WHEN sessions.title_source = 'manual' THEN sessions.title WHEN excluded.title != '' THEN excluded.title ELSE sessions.title END,
			title_source = CASE WHEN sessions.title_source = 'manual' THEN sessions.title_source ELSE excluded.title_source END,
			goal_summary = CASE WHEN excluded.goal_summary != '' THEN excluded.goal_summary ELSE sessions.goal_summary END,
			status = excluded.status,
			last_activity_at = excluded.last_activity_at,
			created_at = CASE WHEN ? = 1 AND excluded.created_at != '' THEN excluded.created_at ELSE sessions.created_at END,
			updated_at = excluded.updated_at,
			resume_command = excluded.resume_command,
			attach_command = excluded.attach_command,
			raw_path = excluded.raw_path`,
		sess.ID, sess.Source, sess.NativeID, projectID, sess.ProjectPath, sess.Agent, sess.Title, titleSource,
		sess.GoalSummary, sess.Status, boolToInt(sess.Pinned), sess.Tags, encodeTime(sess.LastActivityAt), encodeTime(sess.CreatedAt),
		encodeTime(sess.UpdatedAt), sess.ResumeCommand, sess.AttachCommand, sess.RawPath, boolToInt(replaceCreatedAt))
	return err
}

func (s *Store) UpsertRuntime(ctx context.Context, rt core.RuntimeInstance) error {
	rt = normalizeRuntime(rt)
	if strings.TrimSpace(rt.Backend) == "" {
		return fmt.Errorf("runtime backend is required")
	}
	return upsertRuntime(ctx, s.db, rt)
}

// ReplaceRuntimes atomically replaces one backend's complete live snapshot.
// Callers must only use it after an authoritative provider scan succeeds.
func (s *Store) ReplaceRuntimes(ctx context.Context, backend string, runtimes []core.RuntimeInstance) error {
	backend = strings.TrimSpace(backend)
	if backend == "" {
		return fmt.Errorf("runtime backend is required")
	}
	normalized := make([]core.RuntimeInstance, len(runtimes))
	for index, runtime := range runtimes {
		if runtime.Backend == "" {
			runtime.Backend = backend
		}
		if runtime.Backend != backend {
			return fmt.Errorf("runtime backend %q does not match snapshot backend %q", runtime.Backend, backend)
		}
		normalized[index] = normalizeRuntime(runtime)
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback() //nolint:errcheck

	if _, err := tx.ExecContext(ctx, `DELETE FROM runtime_instances WHERE backend = ?`, backend); err != nil {
		return err
	}
	for _, runtime := range normalized {
		if err := upsertRuntime(ctx, tx, runtime); err != nil {
			return err
		}
	}
	return tx.Commit()
}

type dbExecer interface {
	ExecContext(context.Context, string, ...any) (sql.Result, error)
}

func upsertRuntime(ctx context.Context, executor dbExecer, rt core.RuntimeInstance) error {
	_, err := executor.ExecContext(ctx, `INSERT INTO runtime_instances
		(id, session_id, backend, native_id, surface, project_path, command, status, attach_command, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(backend, native_id) DO UPDATE SET
			session_id = excluded.session_id,
			surface = excluded.surface,
			project_path = excluded.project_path,
			command = excluded.command,
			status = excluded.status,
			attach_command = excluded.attach_command,
			updated_at = excluded.updated_at`,
		rt.ID, rt.SessionID, rt.Backend, rt.NativeID, rt.Surface, rt.ProjectPath, rt.Command,
		rt.Status, rt.AttachCommand, encodeTime(rt.UpdatedAt))
	return err
}

func normalizeRuntime(rt core.RuntimeInstance) core.RuntimeInstance {
	if rt.UpdatedAt.IsZero() {
		rt.UpdatedAt = time.Now().UTC()
	}
	if rt.Status == "" {
		rt.Status = core.StatusUnknown
	}
	return rt
}

func (s *Store) ListSessions(ctx context.Context, filter Filter) ([]core.Session, error) {
	var where []string
	var args []any
	if filter.Source != "" {
		where = append(where, "source = ?")
		args = append(args, filter.Source)
	}
	if filter.Status != "" {
		where = append(where, "status = ?")
		args = append(args, filter.Status)
	}
	if filter.PinnedOnly {
		where = append(where, "pinned = 1")
	}
	if filter.Tag != "" {
		where = append(where, "(',' || tags || ',') LIKE ?")
		args = append(args, "%,"+filter.Tag+",%")
	}
	if filter.Query != "" {
		q := "%" + filter.Query + "%"
		where = append(where, "(title LIKE ? OR goal_summary LIKE ? OR project_path LIKE ? OR native_id LIKE ? OR agent LIKE ? OR tags LIKE ?)")
		args = append(args, q, q, q, q, q, q)
	}
	query := `SELECT id, source, native_id, project_path, agent, title, title_source, goal_summary, status,
		pinned, tags,
		last_activity_at, created_at, updated_at, resume_command, attach_command, raw_path FROM sessions`
	if len(where) > 0 {
		query += " WHERE " + strings.Join(where, " AND ")
	}
	query += orderBy(filter.Sort)
	if filter.Limit > 0 {
		query += " LIMIT ?"
		args = append(args, filter.Limit)
	}
	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var sessions []core.Session
	for rows.Next() {
		sess, err := scanSession(rows)
		if err != nil {
			return nil, err
		}
		sessions = append(sessions, sess)
	}
	return sessions, rows.Err()
}

func (s *Store) GetSession(ctx context.Context, ref string) (core.Session, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id, source, native_id, project_path, agent, title, title_source, goal_summary, status,
		pinned, tags, last_activity_at, created_at, updated_at, resume_command, attach_command, raw_path
		FROM sessions
		WHERE id = ? OR id LIKE ? OR native_id = ? OR native_id LIKE ?
		ORDER BY CASE WHEN id = ? THEN 0 ELSE 1 END, length(id)
		LIMIT 2`,
		ref, ref+"%", ref, ref+"%", ref)
	if err != nil {
		return core.Session{}, err
	}
	defer rows.Close()
	var found []core.Session
	for rows.Next() {
		sess, err := scanSession(rows)
		if err != nil {
			return core.Session{}, err
		}
		found = append(found, sess)
	}
	if err := rows.Err(); err != nil {
		return core.Session{}, err
	}
	if len(found) == 0 {
		return core.Session{}, sql.ErrNoRows
	}
	if len(found) > 1 && found[0].ID != ref {
		return core.Session{}, fmt.Errorf("ambiguous session reference %q", ref)
	}
	return found[0], nil
}

func (s *Store) RuntimeForSession(ctx context.Context, sessionID, preferredBackend string) (core.RuntimeInstance, error) {
	query := `SELECT id, session_id, backend, native_id, surface, project_path, command, status, attach_command, updated_at
		FROM runtime_instances WHERE session_id = ?`
	var args []any
	args = append(args, sessionID)
	if preferredBackend != "" {
		query += ` ORDER BY CASE WHEN backend = ? THEN 0 ELSE 1 END, updated_at DESC`
		args = append(args, preferredBackend)
	} else {
		query += ` ORDER BY updated_at DESC`
	}
	query += ` LIMIT 1`
	row := s.db.QueryRowContext(ctx, query, args...)
	return scanRuntime(row)
}

func (s *Store) UpdateSummary(ctx context.Context, sessionID, provider, title, summary, status string, confidence float64) error {
	now := time.Now().UTC()
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback() //nolint:errcheck
	_, err = tx.ExecContext(ctx, `INSERT INTO summaries
		(session_id, provider, title, goal_summary, confidence, created_at)
		VALUES (?, ?, ?, ?, ?, ?)`,
		sessionID, provider, title, summary, confidence, encodeTime(now))
	if err != nil {
		return err
	}
	parts := []string{"title = CASE WHEN title_source = 'manual' THEN title ELSE ? END", "title_source = CASE WHEN title_source = 'manual' THEN title_source ELSE ? END", "goal_summary = ?", "updated_at = ?"}
	args := []any{title, provider, summary, encodeTime(now)}
	if status != "" {
		parts = append(parts, "status = ?")
		args = append(args, status)
	}
	args = append(args, sessionID)
	_, err = tx.ExecContext(ctx, `UPDATE sessions SET `+strings.Join(parts, ", ")+` WHERE id = ?`, args...)
	if err != nil {
		return err
	}
	return tx.Commit()
}

func (s *Store) UpdateTitle(ctx context.Context, sessionID, title string) error {
	_, err := s.db.ExecContext(ctx, `UPDATE sessions SET title = ?, title_source = 'manual', updated_at = ? WHERE id = ?`,
		title, encodeTime(time.Now().UTC()), sessionID)
	return err
}

func (s *Store) UpdateStatus(ctx context.Context, sessionID, status string) error {
	_, err := s.db.ExecContext(ctx, `UPDATE sessions SET status = ?, updated_at = ? WHERE id = ?`,
		status, encodeTime(time.Now().UTC()), sessionID)
	return err
}

func (s *Store) UpdatePinned(ctx context.Context, sessionID string, pinned bool) error {
	_, err := s.db.ExecContext(ctx, `UPDATE sessions SET pinned = ?, updated_at = ? WHERE id = ?`,
		boolToInt(pinned), encodeTime(time.Now().UTC()), sessionID)
	return err
}

func (s *Store) UpdateTags(ctx context.Context, sessionID, tags string) error {
	_, err := s.db.ExecContext(ctx, `UPDATE sessions SET tags = ?, updated_at = ? WHERE id = ?`,
		normalizeTags(tags), encodeTime(time.Now().UTC()), sessionID)
	return err
}

func (s *Store) CountSessions(ctx context.Context) (int, error) {
	var count int
	err := s.db.QueryRowContext(ctx, `SELECT count(*) FROM sessions`).Scan(&count)
	return count, err
}

func (s *Store) upsertProject(ctx context.Context, id, path string) error {
	return upsertProject(ctx, s.db, id, path)
}

func upsertProject(ctx context.Context, executor dbExecer, id, path string) error {
	now := time.Now().UTC().Format(time.RFC3339Nano)
	_, err := executor.ExecContext(ctx, `INSERT INTO projects (id, path, name, updated_at)
		VALUES (?, ?, ?, ?)
		ON CONFLICT(path) DO UPDATE SET name = excluded.name, updated_at = excluded.updated_at`,
		id, path, core.ProjectName(path), now)
	return err
}

func (s *Store) columnExists(ctx context.Context, table, column string) (bool, error) {
	rows, err := s.db.QueryContext(ctx, `PRAGMA table_info(`+table+`)`)
	if err != nil {
		return false, err
	}
	defer rows.Close()
	for rows.Next() {
		var cid int
		var name, typ string
		var notNull int
		var defaultValue any
		var pk int
		if err := rows.Scan(&cid, &name, &typ, &notNull, &defaultValue, &pk); err != nil {
			return false, err
		}
		if name == column {
			return true, nil
		}
	}
	return false, rows.Err()
}

type scanner interface {
	Scan(dest ...any) error
}

func scanSession(row scanner) (core.Session, error) {
	var sess core.Session
	var last, created, updated sql.NullString
	var pinned int
	err := row.Scan(&sess.ID, &sess.Source, &sess.NativeID, &sess.ProjectPath, &sess.Agent,
		&sess.Title, &sess.TitleSource, &sess.GoalSummary, &sess.Status, &pinned, &sess.Tags, &last, &created, &updated,
		&sess.ResumeCommand, &sess.AttachCommand, &sess.RawPath)
	if err != nil {
		return core.Session{}, err
	}
	sess.Pinned = pinned == 1
	sess.LastActivityAt = decodeTime(last.String)
	sess.CreatedAt = decodeTime(created.String)
	sess.UpdatedAt = decodeTime(updated.String)
	return sess, nil
}

func orderBy(sort string) string {
	switch sort {
	case "title":
		return " ORDER BY title COLLATE NOCASE ASC"
	case "project":
		return " ORDER BY project_path COLLATE NOCASE ASC, COALESCE(last_activity_at, updated_at) DESC"
	case "source":
		return " ORDER BY source ASC, COALESCE(last_activity_at, updated_at) DESC"
	case "status":
		return " ORDER BY status ASC, COALESCE(last_activity_at, updated_at) DESC"
	default:
		return " ORDER BY pinned DESC, COALESCE(last_activity_at, updated_at) DESC, updated_at DESC"
	}
}

func boolToInt(v bool) int {
	if v {
		return 1
	}
	return 0
}

func normalizeTags(tags string) string {
	var out []string
	seen := make(map[string]bool)
	for _, tag := range strings.Split(tags, ",") {
		tag = strings.TrimSpace(strings.ToLower(tag))
		if tag == "" || seen[tag] {
			continue
		}
		seen[tag] = true
		out = append(out, tag)
	}
	return strings.Join(out, ",")
}

func scanRuntime(row scanner) (core.RuntimeInstance, error) {
	var rt core.RuntimeInstance
	var updated sql.NullString
	err := row.Scan(&rt.ID, &rt.SessionID, &rt.Backend, &rt.NativeID, &rt.Surface,
		&rt.ProjectPath, &rt.Command, &rt.Status, &rt.AttachCommand, &updated)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return core.RuntimeInstance{}, err
		}
		return core.RuntimeInstance{}, err
	}
	rt.UpdatedAt = decodeTime(updated.String)
	return rt, nil
}

func encodeTime(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.UTC().Format(time.RFC3339Nano)
}

func decodeTime(s string) time.Time {
	if s == "" {
		return time.Time{}
	}
	t, _ := time.Parse(time.RFC3339Nano, s)
	return t
}

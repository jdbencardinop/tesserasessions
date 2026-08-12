package adapters

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	_ "modernc.org/sqlite"
)

type readOnlySQLite struct {
	*sql.DB
	snapshotDir string
}

func (db *readOnlySQLite) Close() error {
	return errors.Join(db.DB.Close(), os.RemoveAll(db.snapshotDir))
}

func openReadOnlySQLite(ctx context.Context, path string) (*readOnlySQLite, error) {
	absolute, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}
	absolute, err = filepath.EvalSymlinks(absolute)
	if err != nil {
		return nil, err
	}
	snapshotDir, err := os.MkdirTemp("", "tss-sqlite-snapshot-")
	if err != nil {
		return nil, err
	}
	snapshotPath := filepath.Join(snapshotDir, "source.db")
	if err := copyStableSQLiteSnapshot(ctx, absolute, snapshotPath); err != nil {
		_ = os.RemoveAll(snapshotDir)
		return nil, err
	}

	uri, err := url.Parse(sqliteFileURI(snapshotPath))
	if err != nil {
		_ = os.RemoveAll(snapshotDir)
		return nil, err
	}
	query := uri.Query()
	query.Set("mode", "rw")
	query.Add("_pragma", "busy_timeout(5000)")
	query.Add("_pragma", "query_only(1)")
	uri.RawQuery = query.Encode()

	db, err := sql.Open("sqlite", uri.String())
	if err != nil {
		_ = os.RemoveAll(snapshotDir)
		return nil, err
	}
	db.SetMaxOpenConns(1)
	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		_ = os.RemoveAll(snapshotDir)
		return nil, fmt.Errorf("opening SQLite read-only: %w", err)
	}
	return &readOnlySQLite{DB: db, snapshotDir: snapshotDir}, nil
}

type sqliteSnapshotState struct {
	main    os.FileInfo
	wal     os.FileInfo
	journal os.FileInfo
}

func copyStableSQLiteSnapshot(ctx context.Context, source, destination string) error {
	const attempts = 3
	for attempt := 0; attempt < attempts; attempt++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		before, err := readSQLiteSnapshotState(source)
		if err != nil {
			return err
		}
		_ = os.Remove(destination)
		_ = os.Remove(destination + "-wal")
		_ = os.Remove(destination + "-journal")
		if err := copySQLiteFile(source, destination); err != nil {
			return err
		}
		if before.wal != nil {
			if err := copySQLiteFile(source+"-wal", destination+"-wal"); err != nil {
				continue
			}
		}
		if before.journal != nil {
			if err := copySQLiteFile(source+"-journal", destination+"-journal"); err != nil {
				continue
			}
		}
		after, err := readSQLiteSnapshotState(source)
		if err != nil {
			return err
		}
		if sameSQLiteSnapshotState(before, after) {
			return nil
		}
	}
	return fmt.Errorf("SQLite database changed while creating a read-only snapshot")
}

func readSQLiteSnapshotState(path string) (sqliteSnapshotState, error) {
	main, err := os.Stat(path)
	if err != nil {
		return sqliteSnapshotState{}, err
	}
	wal, err := os.Stat(path + "-wal")
	if err != nil && !os.IsNotExist(err) {
		return sqliteSnapshotState{}, err
	}
	if os.IsNotExist(err) {
		wal = nil
	}
	journal, err := os.Stat(path + "-journal")
	if err != nil && !os.IsNotExist(err) {
		return sqliteSnapshotState{}, err
	}
	if os.IsNotExist(err) {
		journal = nil
	}
	return sqliteSnapshotState{main: main, wal: wal, journal: journal}, nil
}

func sameSQLiteSnapshotState(left, right sqliteSnapshotState) bool {
	return sameSQLiteFileState(left.main, right.main) &&
		sameSQLiteFileState(left.wal, right.wal) &&
		sameSQLiteFileState(left.journal, right.journal)
}

func sameSQLiteFileState(left, right os.FileInfo) bool {
	if left == nil || right == nil {
		return left == nil && right == nil
	}
	return left.Size() == right.Size() &&
		left.ModTime() == right.ModTime()
}

func copySQLiteFile(source, destination string) (err error) {
	input, err := os.Open(source)
	if err != nil {
		return err
	}
	defer input.Close() //nolint:errcheck
	output, err := os.OpenFile(destination, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o600)
	if err != nil {
		return err
	}
	defer func() {
		err = errors.Join(err, output.Close())
	}()
	if _, err := io.Copy(output, input); err != nil {
		return err
	}
	return output.Sync()
}

func sqliteFileURI(path string) string {
	return sqliteFileURIForOS(path, runtime.GOOS)
}

func sqliteFileURIForOS(path, goos string) string {
	if goos != "windows" {
		return (&url.URL{Scheme: "file", Path: path}).String()
	}
	slashed := strings.ReplaceAll(path, `\`, "/")
	if strings.HasPrefix(slashed, "//") {
		withoutPrefix := strings.TrimPrefix(slashed, "//")
		host, rest, found := strings.Cut(withoutPrefix, "/")
		if found {
			return (&url.URL{Scheme: "file", Host: host, Path: "/" + rest}).String()
		}
	}
	if len(slashed) >= 2 && slashed[1] == ':' {
		slashed = "/" + slashed
	}
	return (&url.URL{Scheme: "file", Path: slashed}).String()
}

func unixSeconds(value float64) time.Time {
	if value <= 0 {
		return time.Time{}
	}
	seconds := int64(value)
	nanoseconds := int64((value - float64(seconds)) * float64(time.Second))
	return time.Unix(seconds, nanoseconds).UTC()
}

func unixMilliseconds(value int64) time.Time {
	if value <= 0 {
		return time.Time{}
	}
	return time.UnixMilli(value).UTC()
}

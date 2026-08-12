package adapters

import (
	"context"
	"database/sql"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/jdbencardinop/tesserasessions/internal/core"
)

type OpenCodeScanner struct {
	Database string
}

func (s OpenCodeScanner) Name() string {
	return "opencode"
}

func (s OpenCodeScanner) Scan(ctx context.Context) core.ScanResult {
	result := core.ScanResult{Source: s.Name()}
	if s.Database == ":memory:" || !core.ExistingFile(s.Database) {
		result.Skipped = true
		result.Message = "OpenCode database not found"
		return result
	}
	db, err := openReadOnlySQLite(ctx, s.Database)
	if err != nil {
		result.Err = err
		return result
	}
	defer db.Close() //nolint:errcheck

	var orphaned int
	if err := db.QueryRowContext(ctx, `
		SELECT COUNT(*)
		FROM "session" s
		LEFT JOIN project p ON p.id = s.project_id
		WHERE p.id IS NULL
	`).Scan(&orphaned); err != nil {
		result.Err = fmt.Errorf("checking OpenCode session relationships: %w", err)
		return result
	}
	if orphaned > 0 {
		result.Err = fmt.Errorf("OpenCode database has %d session(s) without a project", orphaned)
		return result
	}

	rows, err := db.QueryContext(ctx, `
		SELECT
			s.id,
			COALESCE(s.slug, ''),
			COALESCE(s.title, ''),
			COALESCE(s.directory, ''),
			s.path,
			s.time_created,
			s.time_updated,
			COALESCE(p.worktree, ''),
			COALESCE(p.name, '')
		FROM "session" s
		JOIN project p ON p.id = s.project_id
		WHERE s.time_archived IS NULL
		  AND s.parent_id IS NULL
		ORDER BY s.time_updated DESC
	`)
	if err != nil {
		result.Err = fmt.Errorf("querying OpenCode sessions: %w", err)
		return result
	}
	defer rows.Close()

	for rows.Next() {
		if ctx.Err() != nil {
			result.Err = ctx.Err()
			return result
		}
		var (
			id, slug, title, directory, worktree, projectName string
			relativePath                                      sql.NullString
			createdMS, updatedMS                              int64
		)
		if err := rows.Scan(
			&id,
			&slug,
			&title,
			&directory,
			&relativePath,
			&createdMS,
			&updatedMS,
			&worktree,
			&projectName,
		); err != nil {
			result.Err = err
			return result
		}
		id = strings.TrimSpace(id)
		if id == "" {
			result.Err = fmt.Errorf("OpenCode session has empty id")
			return result
		}
		projectPath := openCodeProjectPath(directory, worktree, relativePath)
		displayTitle := strings.TrimSpace(title)
		if displayTitle == "" {
			displayTitle = strings.TrimSpace(slug)
		}
		if displayTitle == "" {
			displayTitle = firstNonEmpty(strings.TrimSpace(projectName), "OpenCode: "+core.ProjectName(projectPath))
		}

		session := core.NewSession("opencode", id)
		session.ProjectPath = projectPath
		session.Agent = "opencode"
		session.Title = core.Truncate(displayTitle, 120)
		session.Status = core.StatusUnknown
		session.CreatedAt = unixMilliseconds(createdMS)
		session.LastActivityAt = unixMilliseconds(updatedMS)
		resume := commandWithEnv(
			"OPENCODE_DB",
			s.Database,
			"opencode --session "+core.ShellQuote(id),
		)
		session.ResumeCommand = resumeInDirectory(projectPath, resume)
		session.RawPath = s.Database
		result.Sessions = append(result.Sessions, session)
	}
	if err := rows.Err(); err != nil {
		result.Err = err
		return result
	}
	result.SessionSnapshotComplete = true
	return result
}

func openCodeProjectPath(directory, worktree string, relativePath sql.NullString) string {
	directory = strings.TrimSpace(directory)
	if directory != "" {
		return filepath.Clean(directory)
	}
	worktree = strings.TrimSpace(worktree)
	if worktree == "" {
		return ""
	}
	if relativePath.Valid && strings.TrimSpace(relativePath.String) != "" {
		return filepath.Clean(filepath.Join(worktree, filepath.FromSlash(relativePath.String)))
	}
	return filepath.Clean(worktree)
}

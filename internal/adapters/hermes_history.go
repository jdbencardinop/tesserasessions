package adapters

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/jdbencardinop/tesserasessions/internal/core"
)

type HermesHistoryScanner struct {
	Database string
}

func (s HermesHistoryScanner) Name() string {
	return "hermes"
}

func (s HermesHistoryScanner) Scan(ctx context.Context) core.ScanResult {
	result := core.ScanResult{Source: s.Name()}
	if !core.ExistingFile(s.Database) {
		result.Skipped = true
		result.Message = "Hermes state database not found"
		return result
	}
	db, err := openReadOnlySQLite(ctx, s.Database)
	if err != nil {
		result.Err = err
		return result
	}
	defer db.Close() //nolint:errcheck

	rows, err := db.QueryContext(ctx, `
		WITH RECURSIVE
		roots AS (
			SELECT id, title, cwd, git_repo_root, started_at
			FROM sessions
			WHERE source = 'cli'
			  AND parent_session_id IS NULL
			  AND COALESCE(archived, 0) = 0
			  AND json_extract(COALESCE(model_config, '{}'), '$._branched_from') IS NULL
			  AND json_extract(COALESCE(model_config, '{}'), '$._delegate_from') IS NULL
		),
		lineage(root_id, id, depth) AS (
			SELECT id, id, 0 FROM roots
			UNION ALL
			SELECT lineage.root_id, child.id, lineage.depth + 1
			FROM lineage
			JOIN sessions parent ON parent.id = lineage.id
			JOIN sessions child ON child.id = (
				SELECT candidate.id
				FROM sessions candidate
				WHERE candidate.parent_session_id = parent.id
				  AND parent.end_reason = 'compression'
				  AND json_extract(COALESCE(candidate.model_config, '{}'), '$._branched_from') IS NULL
				  AND json_extract(COALESCE(candidate.model_config, '{}'), '$._delegate_from') IS NULL
				  AND COALESCE(candidate.source, '') != 'tool'
				  AND COALESCE(candidate.archived, 0) = 0
				ORDER BY
				  CASE
				    WHEN candidate.end_reason = 'compression' THEN 0
				    WHEN candidate.ended_at IS NULL THEN 1
				    ELSE 2
				  END,
				  MAX(
				    COALESCE(candidate.last_activity_at, 0),
				    COALESCE((
				      SELECT MAX(message.timestamp)
				      FROM messages message
				      WHERE message.session_id = candidate.id
				        AND COALESCE(message.active, 1) = 1
				    ), 0),
				    candidate.started_at
				  ) DESC,
				  candidate.started_at DESC,
				  candidate.id DESC
				LIMIT 1
			)
			WHERE lineage.depth < 100
		),
		ranked AS (
			SELECT
				root_id,
				id,
				ROW_NUMBER() OVER (PARTITION BY root_id ORDER BY depth DESC, id DESC) AS rank
			FROM lineage
		)
		SELECT
			root.id,
			COALESCE(NULLIF(tip.title, ''), NULLIF(root.title, ''), ''),
			COALESCE(NULLIF(tip.cwd, ''), NULLIF(root.cwd, ''), ''),
			COALESCE(NULLIF(tip.git_repo_root, ''), NULLIF(root.git_repo_root, ''), ''),
			root.started_at,
			MAX(
			  COALESCE(tip.last_activity_at, 0),
			  COALESCE((
			    SELECT MAX(message.timestamp)
			    FROM messages message
			    WHERE message.session_id = tip.id
			      AND COALESCE(message.active, 1) = 1
			  ), 0),
			  tip.started_at
			) AS tip_last_active
		FROM roots root
		JOIN ranked ON ranked.root_id = root.id AND ranked.rank = 1
		JOIN sessions tip ON tip.id = ranked.id
		ORDER BY COALESCE(tip.last_activity_at, tip.ended_at, tip.started_at, root.started_at) DESC
	`)
	if err != nil {
		result.Err = fmt.Errorf("querying Hermes sessions: %w", err)
		return result
	}
	defer rows.Close()

	for rows.Next() {
		if ctx.Err() != nil {
			result.Err = ctx.Err()
			return result
		}
		var (
			id, title, cwd, repoRoot   string
			rootStarted, tipLastActive float64
		)
		if err := rows.Scan(
			&id,
			&title,
			&cwd,
			&repoRoot,
			&rootStarted,
			&tipLastActive,
		); err != nil {
			result.Err = err
			return result
		}
		id = strings.TrimSpace(id)
		if id == "" {
			result.Err = fmt.Errorf("Hermes session has empty id")
			return result
		}
		projectPath := firstNonEmpty(strings.TrimSpace(cwd), strings.TrimSpace(repoRoot))
		displayTitle := strings.TrimSpace(title)
		if displayTitle == "" {
			displayTitle = "Hermes: " + core.ProjectName(projectPath)
		}
		created := unixSeconds(rootStarted)
		updated := unixSeconds(tipLastActive)
		if updated.IsZero() {
			updated = created
		}

		session := core.NewSession("hermes", id)
		session.ProjectPath = projectPath
		session.Agent = "hermes"
		session.Title = core.Truncate(displayTitle, 120)
		session.Status = core.StatusUnknown
		session.CreatedAt = created
		session.LastActivityAt = updated
		resume := hermesResumeCommand(s.Database, id)
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

func hermesResumeCommand(database, sessionID string) string {
	home := filepath.Dir(database)
	command := "hermes --resume " + core.ShellQuote(sessionID)
	if filepath.Base(filepath.Dir(home)) != "profiles" {
		command = commandWithEnv("HERMES_S6_SUPERVISED_CHILD", "1", command)
	}
	return commandWithEnv("HERMES_HOME", home, command)
}

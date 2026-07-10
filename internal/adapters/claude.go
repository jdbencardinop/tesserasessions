package adapters

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/jdbencardinop/tesserasessions/internal/core"
)

type ClaudeScanner struct {
	Root string
}

func (s ClaudeScanner) Name() string {
	return "claude"
}

func (s ClaudeScanner) Scan(ctx context.Context) core.ScanResult {
	result := core.ScanResult{Source: s.Name()}
	if !core.ExistingDir(s.Root) {
		result.Skipped = true
		result.Message = "Claude projects directory not found"
		return result
	}
	entries, err := os.ReadDir(s.Root)
	if err != nil {
		result.Err = err
		return result
	}
	for _, entry := range entries {
		if ctx.Err() != nil {
			result.Err = ctx.Err()
			return result
		}
		if !entry.IsDir() {
			continue
		}
		projectDir := filepath.Join(s.Root, entry.Name())
		projectPath := core.DecodeClaudeProjectDir(entry.Name())
		_ = filepath.WalkDir(projectDir, func(path string, d fs.DirEntry, err error) error {
			if err != nil || ctx.Err() != nil {
				return nil
			}
			if d.IsDir() || filepath.Ext(path) != ".jsonl" {
				return nil
			}
			info, statErr := d.Info()
			if statErr != nil {
				return nil
			}
			nativeID := strings.TrimSuffix(filepath.Base(path), ".jsonl")
			if nativeID == "" {
				nativeID = path
			}
			meta := readJSONLMetadata(path)
			lastActivity := core.NonZeroTime(meta.LastTimestamp, info.ModTime())
			created := core.NonZeroTime(meta.FirstTimestamp, info.ModTime())
			session := core.NewSession("claude", nativeID)
			session.ProjectPath = projectPath
			session.Agent = "claude"
			session.Title = "Claude: " + core.ProjectName(projectPath)
			session.Status = core.StatusUnknown
			session.LastActivityAt = lastActivity.UTC()
			session.CreatedAt = created.UTC()
			session.ResumeCommand = "cd " + core.ShellQuote(projectPath) + " && claude -c"
			session.RawPath = path
			result.Sessions = append(result.Sessions, session)
			return nil
		})
	}
	return result
}

type jsonlMetadata struct {
	FirstTimestamp time.Time
	LastTimestamp  time.Time
}

func readJSONLMetadata(path string) jsonlMetadata {
	file, err := os.Open(path)
	if err != nil {
		return jsonlMetadata{}
	}
	defer file.Close() //nolint:errcheck
	var meta jsonlMetadata
	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, 0, 64*1024), 2*1024*1024)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		var row map[string]any
		if err := json.Unmarshal([]byte(line), &row); err != nil {
			continue
		}
		ts := parseTimestamp(row)
		if ts.IsZero() {
			continue
		}
		if meta.FirstTimestamp.IsZero() {
			meta.FirstTimestamp = ts
		}
		meta.LastTimestamp = ts
	}
	return meta
}

func parseTimestamp(row map[string]any) time.Time {
	for _, key := range []string{"timestamp", "created_at", "createdAt", "time"} {
		raw, ok := row[key].(string)
		if !ok || raw == "" {
			continue
		}
		for _, layout := range []string{time.RFC3339Nano, time.RFC3339, "2006-01-02 15:04:05"} {
			t, err := time.Parse(layout, raw)
			if err == nil {
				return t.UTC()
			}
		}
	}
	return time.Time{}
}

func readTextCandidatesFromJSONL(path string, limit int) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close() //nolint:errcheck
	var out []string
	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, 0, 64*1024), 2*1024*1024)
	for scanner.Scan() {
		if limit > 0 && len(out) >= limit {
			break
		}
		var row any
		if err := json.Unmarshal(scanner.Bytes(), &row); err != nil {
			continue
		}
		for _, text := range extractText(row) {
			text = cleanCandidate(text)
			if text != "" {
				out = append(out, text)
				if limit > 0 && len(out) >= limit {
					break
				}
			}
		}
	}
	if err := scanner.Err(); err != nil && !errors.Is(err, bufio.ErrTooLong) {
		return out, err
	}
	return out, nil
}

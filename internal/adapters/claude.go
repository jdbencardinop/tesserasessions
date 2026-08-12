package adapters

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"io"
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
		fallbackProjectPath := core.DecodeClaudeProjectDir(entry.Name())
		files, err := os.ReadDir(projectDir)
		if err != nil {
			result.Err = err
			return result
		}
		for _, file := range files {
			if ctx.Err() != nil {
				result.Err = ctx.Err()
				return result
			}
			if file.IsDir() || filepath.Ext(file.Name()) != ".jsonl" {
				continue
			}
			path := filepath.Join(projectDir, file.Name())
			info, statErr := file.Info()
			if statErr != nil {
				result.Err = statErr
				return result
			}
			fileSessionID := strings.TrimSuffix(filepath.Base(path), ".jsonl")
			meta, metaErr := readClaudeMetadata(path, fileSessionID)
			if metaErr != nil {
				result.Err = metaErr
				return result
			}
			nativeID := firstNonEmpty(fileSessionID, meta.SessionID)
			if nativeID == "" {
				nativeID = path
			}
			lastActivity := core.NonZeroTime(meta.LastTimestamp, info.ModTime())
			created := core.NonZeroTime(meta.FirstTimestamp, info.ModTime())
			projectPath := firstNonEmpty(meta.CWD, fallbackProjectPath)
			title := meta.Title
			if title == "" {
				title = "Claude: " + core.ProjectName(projectPath)
			}
			session := core.NewSession("claude", nativeID)
			session.ProjectPath = projectPath
			session.Agent = "claude"
			session.Title = core.Truncate(title, 120)
			session.Status = core.StatusUnknown
			session.LastActivityAt = lastActivity.UTC()
			session.CreatedAt = created.UTC()
			resume := commandWithEnv(
				"CLAUDE_CONFIG_DIR",
				filepath.Dir(s.Root),
				"claude --resume "+core.ShellQuote(nativeID),
			)
			session.ResumeCommand = resumeInDirectory(projectPath, resume)
			session.RawPath = path
			result.Sessions = append(result.Sessions, session)
		}
	}
	result.SessionSnapshotComplete = true
	return result
}

type claudeMetadata struct {
	SessionID      string
	CWD            string
	GitBranch      string
	Title          string
	FirstTimestamp time.Time
	LastTimestamp  time.Time
}

func readClaudeMetadata(path, expectedSessionID string) (claudeMetadata, error) {
	file, err := os.Open(path)
	if err != nil {
		return claudeMetadata{}, err
	}
	defer file.Close() //nolint:errcheck
	var meta claudeMetadata
	titleRank := 0
	reader := bufio.NewReaderSize(file, 64*1024)
	for {
		line, err := readBoundedLine(reader, 2*1024*1024)
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return meta, err
		}
		if len(line) == 0 {
			continue
		}
		var row map[string]any
		if err := json.Unmarshal(line, &row); err != nil {
			continue
		}
		rowSessionID := stringValue(row["sessionId"])
		if expectedSessionID != "" && rowSessionID != "" && rowSessionID != expectedSessionID {
			continue
		}
		if meta.SessionID == "" && rowSessionID != "" {
			meta.SessionID = rowSessionID
		}
		if value := stringValue(row["cwd"]); value != "" && meta.CWD == "" {
			meta.CWD = value
		}
		if value := stringValue(row["gitBranch"]); value != "" && meta.GitBranch == "" {
			meta.GitBranch = value
		}
		switch stringValue(row["type"]) {
		case "custom-title":
			setClaudeTitle(&meta, &titleRank, stringValue(row["customTitle"]), 5)
		case "ai-title":
			setClaudeTitle(&meta, &titleRank, stringValue(row["aiTitle"]), 4)
		case "agent-name":
			setClaudeTitle(&meta, &titleRank, stringValue(row["agentName"]), 3)
		case "summary":
			setClaudeTitle(&meta, &titleRank, stringValue(row["summary"]), 2)
		default:
			setClaudeTitle(&meta, &titleRank, stringValue(row["slug"]), 1)
		}
		ts := parseTimestamp(row)
		if ts.IsZero() {
			continue
		}
		if meta.FirstTimestamp.IsZero() || ts.Before(meta.FirstTimestamp) {
			meta.FirstTimestamp = ts
		}
		if meta.LastTimestamp.IsZero() || ts.After(meta.LastTimestamp) {
			meta.LastTimestamp = ts
		}
	}
	return meta, nil
}

func readBoundedLine(reader *bufio.Reader, limit int) ([]byte, error) {
	line := make([]byte, 0, 64*1024)
	oversized := false
	for {
		fragment, prefix, err := reader.ReadLine()
		if err != nil {
			return nil, err
		}
		if !oversized {
			if len(line)+len(fragment) > limit {
				line = nil
				oversized = true
			} else {
				line = append(line, fragment...)
			}
		}
		if !prefix {
			if oversized {
				return nil, nil
			}
			return line, nil
		}
	}
}

func setClaudeTitle(meta *claudeMetadata, currentRank *int, title string, rank int) {
	title = strings.TrimSpace(title)
	if title == "" || rank < *currentRank {
		return
	}
	meta.Title = title
	*currentRank = rank
}

func stringValue(value any) string {
	text, _ := value.(string)
	return strings.TrimSpace(text)
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

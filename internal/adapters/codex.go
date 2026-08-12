package adapters

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/jdbencardinop/tesserasessions/internal/core"
)

const codexMetadataLineLimit = 8 * 1024 * 1024

type CodexScanner struct {
	Home string
}

func (s CodexScanner) Name() string {
	return "codex"
}

func (s CodexScanner) Scan(ctx context.Context) core.ScanResult {
	result := core.ScanResult{Source: s.Name()}
	var messages []string
	sessionsRoot := filepath.Join(s.Home, "sessions")
	if !core.ExistingDir(sessionsRoot) {
		result.Skipped = true
		result.Message = "Codex sessions directory not found"
		return result
	}
	names, err := readCodexSessionNames(filepath.Join(s.Home, "session_index.jsonl"))
	if err != nil {
		names = make(map[string]string)
		messages = append(messages, fmt.Sprintf("Codex session name index unavailable: %v", err))
	}

	candidates := make(map[string]codexCandidate)
	unreadableRollouts := 0
	var firstUnreadable error
	recordUnreadable := func(err error) {
		unreadableRollouts++
		if firstUnreadable == nil {
			firstUnreadable = err
		}
	}
	err = filepath.WalkDir(sessionsRoot, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if ctx.Err() != nil {
			return ctx.Err()
		}
		if entry.IsDir() ||
			!strings.HasPrefix(entry.Name(), "rollout-") ||
			filepath.Ext(entry.Name()) != ".jsonl" {
			return nil
		}
		info, err := entry.Info()
		if err != nil {
			recordUnreadable(fmt.Errorf("%s: reading file metadata: %w", path, err))
			return nil
		}
		if !info.Mode().IsRegular() {
			return nil
		}
		metadata, err := readCodexRolloutMetadata(path)
		if err != nil {
			recordUnreadable(err)
			return nil
		}
		if !metadata.resumable() {
			return nil
		}
		candidate := codexCandidate{metadata: metadata, path: path, updatedAt: info.ModTime().UTC()}
		existing, found := candidates[metadata.ID]
		if !found ||
			candidate.updatedAt.After(existing.updatedAt) ||
			(candidate.updatedAt.Equal(existing.updatedAt) && candidate.path > existing.path) {
			candidates[metadata.ID] = candidate
		}
		return nil
	})
	if err != nil {
		result.Err = fmt.Errorf("scanning Codex sessions: %w", err)
		return result
	}

	result.Sessions = make([]core.Session, 0, len(candidates))
	for _, candidate := range candidates {
		metadata := candidate.metadata
		title := strings.TrimSpace(names[metadata.ID])
		if title == "" {
			title = "Codex: " + core.ProjectName(metadata.CWD)
		}
		session := core.NewSession("codex", metadata.ID)
		session.ProjectPath = metadata.CWD
		session.Agent = "codex"
		session.Title = core.Truncate(title, 120)
		session.Status = core.StatusUnknown
		session.CreatedAt = metadata.CreatedAt
		session.LastActivityAt = candidate.updatedAt
		resume := commandWithEnv("CODEX_HOME", s.Home, "codex resume "+core.ShellQuote(metadata.ID))
		session.ResumeCommand = resumeInDirectory(metadata.CWD, resume)
		session.RawPath = candidate.path
		result.Sessions = append(result.Sessions, session)
	}
	sort.Slice(result.Sessions, func(i, j int) bool {
		if result.Sessions[i].LastActivityAt.Equal(result.Sessions[j].LastActivityAt) {
			return result.Sessions[i].NativeID < result.Sessions[j].NativeID
		}
		return result.Sessions[i].LastActivityAt.After(result.Sessions[j].LastActivityAt)
	})
	if unreadableRollouts > 0 {
		messages = append(messages, fmt.Sprintf(
			"Codex scan incomplete: skipped %d unreadable rollout(s); first error: %v",
			unreadableRollouts,
			firstUnreadable,
		))
	}
	result.Message = strings.Join(messages, "; ")
	result.SessionSnapshotComplete = unreadableRollouts == 0
	return result
}

type codexCandidate struct {
	metadata  codexMetadata
	path      string
	updatedAt time.Time
}

type codexRolloutLine struct {
	Timestamp string          `json:"timestamp"`
	Type      string          `json:"type"`
	Payload   json.RawMessage `json:"payload"`
}

type codexMetadata struct {
	ID           string          `json:"id"`
	ParentThread *string         `json:"parent_thread_id"`
	Timestamp    string          `json:"timestamp"`
	CWD          string          `json:"cwd"`
	Source       json.RawMessage `json:"source"`
	HistoryMode  string          `json:"history_mode"`
	CreatedAt    time.Time       `json:"-"`
}

func (metadata codexMetadata) resumable() bool {
	if metadata.ID == "" || metadata.CWD == "" || !filepath.IsAbs(metadata.CWD) || metadata.CreatedAt.IsZero() {
		return false
	}
	if metadata.ParentThread != nil && strings.TrimSpace(*metadata.ParentThread) != "" {
		return false
	}
	if metadata.HistoryMode != "" &&
		metadata.HistoryMode != "legacy" &&
		metadata.HistoryMode != "paginated" {
		return false
	}
	return codexInteractiveSource(metadata.Source)
}

func readCodexRolloutMetadata(path string) (codexMetadata, error) {
	file, err := os.Open(path)
	if err != nil {
		return codexMetadata{}, err
	}
	defer file.Close() //nolint:errcheck
	reader := bufio.NewReaderSize(file, 64*1024)
	line, err := readBoundedLine(reader, codexMetadataLineLimit)
	if err != nil {
		return codexMetadata{}, fmt.Errorf("%s: reading first rollout line: %w", path, err)
	}
	if len(line) == 0 {
		return codexMetadata{}, fmt.Errorf("%s: empty or oversized first rollout line", path)
	}
	var envelope codexRolloutLine
	if err := json.Unmarshal(line, &envelope); err != nil {
		return codexMetadata{}, fmt.Errorf("%s: invalid first rollout line: %w", path, err)
	}
	if envelope.Type != "session_meta" {
		return codexMetadata{}, fmt.Errorf("%s: first rollout item is %q, want session_meta", path, envelope.Type)
	}
	var metadata codexMetadata
	if err := json.Unmarshal(envelope.Payload, &metadata); err != nil {
		return codexMetadata{}, fmt.Errorf("%s: invalid session metadata: %w", path, err)
	}
	metadata.ID = strings.TrimSpace(metadata.ID)
	metadata.CWD = filepath.Clean(strings.TrimSpace(metadata.CWD))
	metadata.HistoryMode = strings.ToLower(strings.TrimSpace(metadata.HistoryMode))
	metadata.CreatedAt = parseTimeString(firstNonEmpty(metadata.Timestamp, envelope.Timestamp))
	if metadata.ID == "" {
		return codexMetadata{}, fmt.Errorf("%s: session metadata has empty thread id", path)
	}
	if metadata.CWD == "." || metadata.CWD == "" || !filepath.IsAbs(metadata.CWD) {
		return codexMetadata{}, fmt.Errorf("%s: session metadata has invalid cwd %q", path, metadata.CWD)
	}
	if metadata.CreatedAt.IsZero() {
		return codexMetadata{}, fmt.Errorf("%s: session metadata has invalid timestamp", path)
	}
	return metadata, nil
}

func codexInteractiveSource(raw json.RawMessage) bool {
	if len(raw) == 0 {
		return true // SessionSource defaults to VSCode.
	}
	if string(raw) == "null" {
		return false
	}
	var source string
	if err := json.Unmarshal(raw, &source); err == nil {
		source = strings.ToLower(strings.TrimSpace(source))
		return source == "cli" || source == "vscode"
	}
	var custom map[string]string
	if err := json.Unmarshal(raw, &custom); err != nil {
		return false
	}
	value := strings.ToLower(strings.TrimSpace(custom["custom"]))
	return value == "atlas" || value == "chatgpt"
}

type codexSessionIndexEntry struct {
	ID         string `json:"id"`
	ThreadName string `json:"thread_name"`
	UpdatedAt  string `json:"updated_at"`
}

func readCodexSessionNames(path string) (map[string]string, error) {
	names := make(map[string]string)
	file, err := os.Open(path)
	if os.IsNotExist(err) {
		return names, nil
	}
	if err != nil {
		return nil, fmt.Errorf("opening %s: %w", path, err)
	}
	defer file.Close() //nolint:errcheck
	reader := bufio.NewReaderSize(file, 64*1024)
	for {
		line, err := readBoundedLine(reader, 1024*1024)
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("reading %s: %w", path, err)
		}
		if len(line) == 0 {
			continue
		}
		var entry codexSessionIndexEntry
		if json.Unmarshal(line, &entry) != nil {
			continue
		}
		entry.ID = strings.TrimSpace(entry.ID)
		entry.ThreadName = strings.TrimSpace(entry.ThreadName)
		if entry.ID != "" {
			names[entry.ID] = entry.ThreadName
		}
	}
	return names, nil
}

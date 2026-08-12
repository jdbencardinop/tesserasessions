package adapters

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/jdbencardinop/tesserasessions/internal/core"
)

func TestCodexScannerEnumeratesResumableInteractiveThreads(t *testing.T) {
	home := t.TempDir()
	sessions := filepath.Join(home, "sessions")
	old := time.Unix(100, 0)
	newer := time.Unix(200, 0)

	writeCodexRollout(t, sessions, "rollout-legacy-thread-flat.jsonl", codexPayload(
		"thread-flat", "/repo/flat", "cli", nil, "legacy", "2026-01-01T00:00:00Z",
	), old)
	writeCodexRollout(t, sessions, "2026/01/02/rollout-old-thread-cli.jsonl", codexPayload(
		"thread-cli", "/repo/old", "cli", nil, "legacy", "2026-01-02T00:00:00Z",
	), old)
	newestPath := writeCodexRollout(t, sessions, "2026/01/03/rollout-new-thread-cli_rollout.jsonl", codexPayload(
		"thread-cli", "/repo/new", "cli", nil, "paginated", "2026-01-03T00:00:00Z",
	), newer)
	writeCodexRollout(t, sessions, "2026/01/04/rollout-vscode.jsonl", codexPayload(
		"thread-vscode", "/repo/vscode", nil, nil, "", "2026-01-04T00:00:00Z",
	), old)
	writeCodexRollout(t, sessions, "2026/01/05/rollout-atlas.jsonl", codexPayload(
		"thread-atlas", "/repo/atlas", map[string]string{"custom": "atlas"}, nil, "legacy", "2026-01-05T00:00:00Z",
	), old)
	writeCodexRollout(t, sessions, "2026/01/06/rollout-chatgpt.jsonl", codexPayload(
		"thread-chatgpt", "/repo/chatgpt", map[string]string{"custom": "chatgpt"}, nil, "legacy", "2026-01-06T00:00:00Z",
	), old)
	fork := codexPayload(
		"thread-fork", "/repo/fork", "cli", nil, "legacy", "2026-01-06T12:00:00Z",
	)
	fork["forked_from_id"] = "thread-cli"
	writeCodexRollout(t, sessions, "2026/01/06/rollout-fork.jsonl", fork, old)

	parent := "thread-cli"
	writeCodexRollout(t, sessions, "2026/01/07/rollout-subagent.jsonl", codexPayload(
		"thread-subagent", "/repo/subagent", "cli", &parent, "legacy", "2026-01-07T00:00:00Z",
	), old)
	writeCodexRollout(t, sessions, "2026/01/08/rollout-exec.jsonl", codexPayload(
		"thread-exec", "/repo/exec", "exec", nil, "legacy", "2026-01-08T00:00:00Z",
	), old)
	writeCodexRollout(t, sessions, "2026/01/09/rollout-custom.jsonl", codexPayload(
		"thread-custom", "/repo/custom", map[string]string{"custom": "other"}, nil, "legacy", "2026-01-09T00:00:00Z",
	), old)
	writeCodexRollout(t, sessions, "2026/01/10/rollout-future.jsonl", codexPayload(
		"thread-future", "/repo/future", "cli", nil, "future", "2026-01-10T00:00:00Z",
	), old)
	nullSource := codexPayload(
		"thread-null", "/repo/null", nil, nil, "legacy", "2026-01-11T00:00:00Z",
	)
	nullSource["source"] = nil
	writeCodexRollout(t, sessions, "2026/01/11/rollout-null-source.jsonl", nullSource, old)
	writeCodexRollout(t, sessions, "2026/01/12/rollout-internal.jsonl", codexPayload(
		"thread-internal", "/repo/internal", map[string]any{"internal": "compact"}, nil, "legacy", "2026-01-12T00:00:00Z",
	), old)
	writeCodexRollout(t, sessions, "2026/01/13/rollout-source-subagent.jsonl", codexPayload(
		"thread-source-subagent", "/repo/subagent", map[string]any{"subagent": "review"}, nil, "legacy", "2026-01-13T00:00:00Z",
	), old)
	if err := os.WriteFile(filepath.Join(sessions, "unrelated.jsonl"), []byte("{\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(sessions, "rollout-compressed.jsonl.zst"), []byte("{\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	linkTarget := filepath.Join(home, "private-rollout.jsonl")
	if err := os.WriteFile(linkTarget, []byte("{\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(linkTarget, filepath.Join(sessions, "rollout-linked.jsonl")); err != nil {
		t.Logf("symlink fixture unavailable: %v", err)
	}
	archived := filepath.Join(home, "archived_sessions", "rollout-archived.jsonl")
	if err := os.MkdirAll(filepath.Dir(archived), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(archived, []byte("{\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	index := strings.Join([]string{
		`{"id":"thread-cli","thread_name":"old name","updated_at":"2026-01-02T00:00:00Z"}`,
		`not-json`,
		`{"id":"thread-cli","thread_name":"renamed thread","updated_at":"2026-01-03T00:00:00Z"}`,
		`{"id":"thread-flat","thread_name":"flat thread","updated_at":"2026-01-01T00:00:00Z"}`,
	}, "\n") + "\n"
	if err := os.WriteFile(filepath.Join(home, "session_index.jsonl"), []byte(index), 0o644); err != nil {
		t.Fatal(err)
	}

	result := (CodexScanner{Home: home}).Scan(context.Background())
	if result.Err != nil {
		t.Fatal(result.Err)
	}
	if !result.SessionSnapshotComplete {
		t.Fatal("successful Codex scan must be authoritative")
	}
	if len(result.Sessions) != 6 {
		t.Fatalf("sessions = %d, want 6: %+v", len(result.Sessions), result.Sessions)
	}

	thread := findNativeSession(t, result.Sessions, "thread-cli")
	if thread.ProjectPath != "/repo/new" || thread.Title != "renamed thread" || thread.RawPath != newestPath {
		t.Fatalf("unexpected deduplicated thread: %+v", thread)
	}
	if thread.CreatedAt != time.Date(2026, 1, 3, 0, 0, 0, 0, time.UTC) ||
		!thread.LastActivityAt.Equal(newer.UTC()) {
		t.Fatalf("unexpected timestamps: %+v", thread)
	}
	wantResume := "cd '/repo/new' && CODEX_HOME=" +
		core.ShellQuote(home) + " codex resume 'thread-cli'"
	if thread.ResumeCommand != wantResume {
		t.Fatalf("resume = %q, want %q", thread.ResumeCommand, wantResume)
	}
	if findNativeSession(t, result.Sessions, "thread-flat").Title != "flat thread" {
		t.Fatal("flat layout or session name not indexed")
	}
	findNativeSession(t, result.Sessions, "thread-fork")
	if strings.Contains(fmt.Sprint(result.Sessions), "private-value-not-read") {
		t.Fatal("ignored Codex repository URL leaked into normalized sessions")
	}
}

func TestCodexScannerSkipsMissingStore(t *testing.T) {
	result := (CodexScanner{Home: t.TempDir()}).Scan(context.Background())
	if !result.Skipped || result.Err != nil {
		t.Fatalf("missing store result: %+v", result)
	}
}

func TestCodexScannerTreatsNameIndexAsOptionalEnrichment(t *testing.T) {
	home := t.TempDir()
	writeCodexRollout(t, filepath.Join(home, "sessions"), "rollout-thread.jsonl", codexPayload(
		"thread", "/repo/project", "cli", nil, "legacy", "2026-01-01T00:00:00Z",
	), time.Unix(100, 0))
	if err := os.Mkdir(filepath.Join(home, "session_index.jsonl"), 0o755); err != nil {
		t.Fatal(err)
	}

	result := (CodexScanner{Home: home}).Scan(context.Background())
	if result.Err != nil || !result.SessionSnapshotComplete || len(result.Sessions) != 1 {
		t.Fatalf("optional index blocked scan: %+v", result)
	}
	if result.Message == "" || result.Sessions[0].Title != "Codex: project" {
		t.Fatalf("index warning or fallback title missing: %+v", result)
	}
}

func TestCodexScannerRejectsMalformedAuthoritativeMetadata(t *testing.T) {
	tests := []struct {
		name string
		line string
	}{
		{name: "invalid json", line: "{"},
		{name: "wrong item", line: `{"timestamp":"2026-01-01T00:00:00Z","type":"response_item","payload":{}}`},
		{name: "missing id", line: `{"timestamp":"2026-01-01T00:00:00Z","type":"session_meta","payload":{"cwd":"/repo","timestamp":"2026-01-01T00:00:00Z","source":"cli"}}`},
		{name: "relative cwd", line: `{"timestamp":"2026-01-01T00:00:00Z","type":"session_meta","payload":{"id":"id","cwd":"repo","timestamp":"2026-01-01T00:00:00Z","source":"cli"}}`},
		{name: "invalid timestamp", line: `{"timestamp":"invalid","type":"session_meta","payload":{"id":"id","cwd":"/repo","timestamp":"invalid","source":"cli"}}`},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			home := t.TempDir()
			path := filepath.Join(home, "sessions", "rollout-test.jsonl")
			if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(path, []byte(test.line+"\n"), 0o644); err != nil {
				t.Fatal(err)
			}
			result := (CodexScanner{Home: home}).Scan(context.Background())
			if result.Err != nil || result.SessionSnapshotComplete || result.Message == "" {
				t.Fatalf("malformed rollout did not produce an incomplete warning: %+v", result)
			}
		})
	}
}

func TestCodexRolloutUsesOuterTimestampFallback(t *testing.T) {
	home := t.TempDir()
	path := filepath.Join(home, "sessions", "rollout-fallback.jsonl")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	line := `{"timestamp":"2026-01-01T00:00:00Z","type":"session_meta","payload":{"id":"id","cwd":"/repo","source":"cli"}}`
	if err := os.WriteFile(path, []byte(line+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	result := (CodexScanner{Home: home}).Scan(context.Background())
	if result.Err != nil || len(result.Sessions) != 1 {
		t.Fatalf("outer timestamp fallback failed: %+v", result)
	}
	if got := result.Sessions[0].CreatedAt; got != time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC) {
		t.Fatalf("created at = %s", got)
	}
}

func TestCodexScannerRejectsOversizedMetadataLine(t *testing.T) {
	home := t.TempDir()
	path := filepath.Join(home, "sessions", "rollout-test.jsonl")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	line := `{"timestamp":"2026-01-01T00:00:00Z","type":"session_meta","payload":{"id":"id","cwd":"/repo","timestamp":"2026-01-01T00:00:00Z","source":"cli","base_instructions":"` +
		strings.Repeat("x", codexMetadataLineLimit+1) + `"}}`
	if err := os.WriteFile(path, []byte(line+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	result := (CodexScanner{Home: home}).Scan(context.Background())
	if result.Err != nil || result.SessionSnapshotComplete || result.Message == "" {
		t.Fatalf("oversized metadata did not produce an incomplete warning: %+v", result)
	}
}

func codexPayload(id, cwd string, source any, parent *string, historyMode, timestamp string) map[string]any {
	payload := map[string]any{
		"session_id":       "session-" + id,
		"id":               id,
		"parent_thread_id": parent,
		"timestamp":        timestamp,
		"cwd":              cwd,
		"originator":       "codex-cli",
		"cli_version":      "1.0.0",
		"history_mode":     historyMode,
		"git": map[string]any{
			"branch":         "main",
			"commit_hash":    "abc123",
			"repository_url": "private-value-not-read",
		},
	}
	if source != nil {
		payload["source"] = source
	}
	return payload
}

func writeCodexRollout(t *testing.T, sessionsRoot, relative string, payload map[string]any, modified time.Time) string {
	t.Helper()
	path := filepath.Join(sessionsRoot, filepath.FromSlash(relative))
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	line, err := json.Marshal(map[string]any{
		"timestamp": payload["timestamp"],
		"ordinal":   0,
		"type":      "session_meta",
		"payload":   payload,
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, append(line, '\n'), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Chtimes(path, modified, modified); err != nil {
		t.Fatal(err)
	}
	return path
}

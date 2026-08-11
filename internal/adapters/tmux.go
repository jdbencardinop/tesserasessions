package adapters

import (
	"context"
	"errors"
	"os/exec"
	"strings"
	"time"

	"github.com/jdbencardinop/tesserasessions/internal/core"
)

type TmuxScanner struct{}

func (s TmuxScanner) Name() string {
	return "tmux"
}

func (s TmuxScanner) Scan(ctx context.Context) core.ScanResult {
	result := core.ScanResult{Source: s.Name()}
	if _, err := exec.LookPath("tmux"); err != nil {
		result.Skipped = true
		result.Message = "tmux not found"
		return result
	}
	format := strings.Join([]string{
		"#{session_name}",
		"#{window_index}",
		"#{pane_index}",
		"#{pane_id}",
		"#{pane_current_path}",
		"#{pane_current_command}",
		"#{pane_title}",
		"#{pane_active}",
	}, "\x1f")
	out, err := exec.CommandContext(ctx, "tmux", "list-panes", "-a", "-F", format).CombinedOutput()
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) && isNoTmuxServer(string(out)) {
			result.Skipped = true
			result.SnapshotComplete = true
			result.Message = "tmux server not running"
			return result
		}
		if ctx.Err() != nil {
			result.Err = ctx.Err()
			return result
		}
		result.Err = err
		return result
	}
	result.SnapshotComplete = true
	now := time.Now().UTC()
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		if strings.TrimSpace(line) == "" {
			continue
		}
		parts := strings.Split(line, "\x1f")
		if len(parts) < 8 {
			continue
		}
		sessionName, windowIndex, paneIndex, paneID := parts[0], parts[1], parts[2], parts[3]
		cwd, command, title := parts[4], parts[5], parts[6]
		agent, isAgent := detectAgentCommand(command)
		if !isAgent {
			continue
		}
		nativeID := paneID
		sess := core.NewSession("tmux", nativeID)
		sess.ProjectPath = cwd
		sess.Agent = agent
		sess.Title = "tmux " + sessionName + ":" + windowIndex + "." + paneIndex
		if title != "" && title != command {
			sess.Title += " " + core.Truncate(title, 32)
		}
		sess.Status = inferRuntimeStatus(command)
		sess.LastActivityAt = now
		sess.CreatedAt = now
		sess.AttachCommand = core.TmuxAttachCommand(sessionName)
		result.Sessions = append(result.Sessions, sess)
		rt := core.RuntimeInstance{
			ID:            core.RuntimeID("tmux", nativeID),
			SessionID:     sess.ID,
			Backend:       "tmux",
			NativeID:      nativeID,
			Surface:       sessionName + ":" + windowIndex + "." + paneIndex,
			ProjectPath:   cwd,
			Command:       command,
			Status:        sess.Status,
			AttachCommand: sess.AttachCommand,
			UpdatedAt:     now,
		}
		result.Runtimes = append(result.Runtimes, rt)
	}
	return result
}

func detectAgentCommand(command string) (string, bool) {
	command = strings.ToLower(strings.TrimSpace(command))
	agents := map[string]string{
		"aider":        "aider",
		"amp":          "amp",
		"claude":       "claude",
		"cline":        "cline",
		"codex":        "codex",
		"copilot":      "copilot",
		"cursor-agent": "cursor-agent",
		"devin":        "devin",
		"droid":        "droid",
		"gemini":       "gemini",
		"grok":         "grok",
		"hermes":       "hermes",
		"kilo":         "kilo",
		"kimi":         "kimi",
		"kiro":         "kiro",
		"mastracode":   "mastracode",
		"omp":          "omp",
		"opencode":     "opencode",
		"pi":           "pi",
		"qodercli":     "qodercli",
	}
	agent, ok := agents[command]
	return agent, ok
}

func isNoTmuxServer(output string) bool {
	output = strings.ToLower(output)
	return strings.Contains(output, "no server running") ||
		strings.Contains(output, "failed to connect to server") ||
		strings.Contains(output, "error connecting to")
}

func inferRuntimeStatus(command string) string {
	switch strings.ToLower(command) {
	case "", "bash", "zsh", "fish", "sh", "tmux":
		return core.StatusIdle
	default:
		return core.StatusUnknown
	}
}

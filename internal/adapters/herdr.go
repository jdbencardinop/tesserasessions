package adapters

import (
	"context"
	"encoding/json"
	"os/exec"
	"strings"
	"time"

	"github.com/jdbencardinop/tesserasessions/internal/core"
)

type HerdrScanner struct{}

func (s HerdrScanner) Name() string {
	return "herdr"
}

func (s HerdrScanner) Scan(ctx context.Context) core.ScanResult {
	result := core.ScanResult{Source: s.Name()}
	if _, err := exec.LookPath("herdr"); err != nil {
		result.Skipped = true
		result.Message = "herdr not found"
		return result
	}
	out, err := exec.CommandContext(ctx, "herdr", "agent", "list", "--json").Output()
	if err != nil {
		result.Skipped = true
		result.Message = "herdr agent list unavailable; is the server running?"
		return result
	}
	var payload any
	if err := json.Unmarshal(out, &payload); err != nil {
		result.Err = err
		return result
	}
	now := time.Now().UTC()
	for _, item := range objectList(payload) {
		nativeID := firstString(item, "id", "agent_id", "pane_id", "terminal_id", "name")
		if nativeID == "" {
			continue
		}
		cwd := firstString(item, "cwd", "project_path", "foreground_cwd")
		agent := firstString(item, "agent", "label", "name", "display_agent")
		status := normalizeStatus(firstString(item, "status", "state", "agent_state"))
		if status == "" {
			status = core.StatusUnknown
		}
		title := firstString(item, "title", "name", "label")
		if title == "" {
			title = "Herdr: " + core.Truncate(nativeID, 32)
		}
		session := core.NewSession("herdr", nativeID)
		session.ProjectPath = cwd
		session.Agent = firstNonEmpty(agent, "agent")
		session.Title = title
		session.Status = status
		session.LastActivityAt = now
		session.CreatedAt = now
		session.AttachCommand = "herdr agent attach " + core.ShellQuote(nativeID)
		result.Sessions = append(result.Sessions, session)
		result.Runtimes = append(result.Runtimes, core.RuntimeInstance{
			ID:            core.RuntimeID("herdr", nativeID),
			SessionID:     session.ID,
			Backend:       "herdr",
			NativeID:      nativeID,
			Surface:       nativeID,
			ProjectPath:   cwd,
			Command:       agent,
			Status:        status,
			AttachCommand: session.AttachCommand,
			UpdatedAt:     now,
		})
	}
	return result
}

func objectList(v any) []map[string]any {
	switch x := v.(type) {
	case []any:
		return mapsFromArray(x)
	case map[string]any:
		for _, key := range []string{"agents", "items", "data", "results", "panes"} {
			if arr, ok := x[key].([]any); ok {
				return mapsFromArray(arr)
			}
		}
		return []map[string]any{x}
	default:
		return nil
	}
}

func mapsFromArray(arr []any) []map[string]any {
	var out []map[string]any
	for _, item := range arr {
		if m, ok := item.(map[string]any); ok {
			out = append(out, m)
		}
	}
	return out
}

func firstString(m map[string]any, keys ...string) string {
	for _, key := range keys {
		if v, ok := m[key]; ok {
			switch t := v.(type) {
			case string:
				if strings.TrimSpace(t) != "" {
					return strings.TrimSpace(t)
				}
			case map[string]any:
				if s := firstString(t, "id", "name", "label", "path", "state", "status"); s != "" {
					return s
				}
			}
		}
	}
	return ""
}

func normalizeStatus(status string) string {
	switch strings.ToLower(strings.TrimSpace(status)) {
	case "blocked", "waiting", "needs_attention", "needs-attention":
		return core.StatusNeedsAttention
	case "working", "busy", "running":
		return core.StatusWorking
	case "idle", "ready":
		return core.StatusIdle
	case "done", "complete", "completed":
		return core.StatusDone
	case "stale":
		return core.StatusStale
	default:
		return ""
	}
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}

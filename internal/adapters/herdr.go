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
		if ctx.Err() != nil {
			result.Err = ctx.Err()
			return result
		}
		result.Skipped = true
		result.Message = "herdr agent list unavailable; is the server running?"
		return result
	}
	var payload any
	if err := json.Unmarshal(out, &payload); err != nil {
		result.Err = err
		return result
	}
	result.SnapshotComplete = true
	now := time.Now().UTC()
	for _, item := range objectList(payload) {
		target, paneID := herdrTarget(item)
		if target == "" {
			continue
		}
		cwd := firstString(item, "foreground_cwd", "cwd", "project_path", "workspace_cwd")
		agent := firstNonEmpty(firstString(item, "agent", "kind", "agent_kind", "label", "display_agent"), "agent")
		status := normalizeStatus(firstString(item, "status", "state", "agent_state", "lifecycle_state"))
		if status == "" {
			status = core.StatusUnknown
		}
		title := firstString(item, "title", "display_title", "name", "label")
		if title == "" {
			title = "Herdr: " + core.Truncate(target, 32)
		}
		session := core.NewSession("herdr", target)
		session.ProjectPath = cwd
		session.Agent = agent
		session.Title = title
		session.Status = status
		session.LastActivityAt = now
		session.CreatedAt = now
		session.AttachCommand = "herdr agent attach " + core.ShellQuote(target)
		if ref := agentSessionRef(item); ref != "" {
			session.ResumeCommand = resumeCommand(agent, ref)
			session.RawPath = ref
		}
		result.Sessions = append(result.Sessions, session)
		result.Runtimes = append(result.Runtimes, core.RuntimeInstance{
			ID:            core.RuntimeID("herdr", target),
			SessionID:     session.ID,
			Backend:       "herdr",
			NativeID:      target,
			Surface:       firstNonEmpty(paneID, target),
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
		for _, key := range []string{"agents", "items", "data", "results", "panes", "workspaces", "tabs"} {
			if nested, ok := x[key]; ok {
				if list := objectList(nested); len(list) > 0 {
					return list
				}
			}
		}
		for _, key := range []string{"result", "response"} {
			if nested, ok := x[key]; ok {
				if list := objectList(nested); len(list) > 0 {
					return list
				}
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
			if s := stringFromAny(v); s != "" {
				return s
			}
		}
	}
	return ""
}

func stringFromAny(v any) string {
	switch t := v.(type) {
	case string:
		return strings.TrimSpace(t)
	case map[string]any:
		return firstString(t, "pane_id", "id", "name", "label", "path", "state", "status", "cwd")
	default:
		return ""
	}
}

func herdrTarget(item map[string]any) (target, paneID string) {
	paneID = firstString(item, "pane_id", "paneId", "public_pane_id")
	target = firstString(item, "name", "agent_name", "agent_id")
	if target == "" {
		target = firstString(item, "id")
	}
	if target == "" {
		target = paneID
	}
	if paneID == "" {
		paneID = firstString(item, "terminal_id", "terminalId")
	}
	return target, paneID
}

func agentSessionRef(item map[string]any) string {
	if ref := firstString(item, "agent_session_id", "session_id", "agent_session_path", "session_path"); ref != "" {
		return ref
	}
	if nested, ok := item["agent_session"].(map[string]any); ok {
		return firstString(nested, "id", "session_id", "path", "agent_session_path")
	}
	return ""
}

func resumeCommand(agent, ref string) string {
	switch strings.ToLower(agent) {
	case "claude", "claude code", "claude-code":
		return "claude --resume " + core.ShellQuote(ref)
	case "copilot", "github copilot cli":
		return "copilot --resume=" + core.ShellQuote(ref)
	case "hermes", "hermes agent":
		return "hermes --resume " + core.ShellQuote(ref)
	case "opencode":
		return "opencode --session " + core.ShellQuote(ref)
	case "codex":
		return "codex resume " + core.ShellQuote(ref)
	default:
		return ""
	}
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

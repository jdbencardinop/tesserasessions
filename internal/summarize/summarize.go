package summarize

import (
	"context"
	"os/exec"
	"regexp"
	"strings"

	"github.com/jdbencardinop/tesserasessions/internal/adapters"
	"github.com/jdbencardinop/tesserasessions/internal/core"
)

type Result struct {
	Title      string
	Summary    string
	Status     string
	Confidence float64
}

func Local(ctx context.Context, sess core.Session) Result {
	candidates := meaningfulCandidates(readCandidates(ctx, sess))
	title := sess.Title
	if len(candidates) > 0 {
		title = deriveTitle(candidates[0], sess)
	}
	if title == "" {
		title = sess.Agent + ": " + core.ProjectName(sess.ProjectPath)
	}
	summary := sess.GoalSummary
	if len(candidates) > 0 {
		summary = composeSummary(candidates, sess)
	}
	if summary == "" {
		summary = "Session in " + core.ProjectName(sess.ProjectPath)
	}
	status := inferStatusFromText(strings.Join(candidates, " "))
	return Result{
		Title:      title,
		Summary:    summary,
		Status:     status,
		Confidence: confidence(candidates),
	}
}

func readCandidates(ctx context.Context, sess core.Session) []string {
	if sess.RawPath != "" {
		items, err := adapters.ReadTextCandidates(sess.RawPath, 20)
		if err == nil && len(items) > 0 {
			return items
		}
	}
	if sess.Source == "tmux" && sess.NativeID != "" {
		out, err := exec.CommandContext(ctx, "tmux", "capture-pane", "-p", "-t", sess.NativeID, "-S", "-120").Output()
		if err == nil {
			return linesFromText(string(out), 20)
		}
	}
	return nil
}

func selectSummaryCandidates(candidates []string) []string {
	if len(candidates) <= 2 {
		return candidates
	}
	return []string{candidates[0], candidates[len(candidates)-1]}
}

func meaningfulCandidates(candidates []string) []string {
	var out []string
	for _, candidate := range candidates {
		candidate = strings.TrimSpace(strings.Join(strings.Fields(candidate), " "))
		if !isMeaningful(candidate) {
			continue
		}
		out = append(out, candidate)
		if len(out) >= 12 {
			break
		}
	}
	return out
}

func isMeaningful(text string) bool {
	if len(text) < 12 {
		return false
	}
	lower := strings.ToLower(text)
	for _, needle := range []string{"tool_use_id", "base64", "stack trace", "panic:", "node_modules/", "go: downloading"} {
		if strings.Contains(lower, needle) {
			return false
		}
	}
	return true
}

func deriveTitle(first string, sess core.Session) string {
	title := first
	replacements := []string{
		`(?i)^please\s+`, "",
		`(?i)^can you\s+`, "",
		`(?i)^could you\s+`, "",
		`(?i)^i want to\s+`, "",
		`(?i)^let'?s\s+`, "",
		`(?i)^we need to\s+`, "",
	}
	for i := 0; i < len(replacements); i += 2 {
		title = regexp.MustCompile(replacements[i]).ReplaceAllString(title, replacements[i+1])
	}
	title = strings.Trim(title, " .,:;!?")
	if title == "" {
		return sess.Agent + ": " + core.ProjectName(sess.ProjectPath)
	}
	return core.Truncate(title, 72)
}

func composeSummary(candidates []string, sess core.Session) string {
	first := core.Truncate(candidates[0], 180)
	if len(candidates) == 1 {
		return "Goal: " + first
	}
	latest := core.Truncate(candidates[len(candidates)-1], 180)
	if first == latest {
		return "Goal: " + first
	}
	project := core.ProjectName(sess.ProjectPath)
	if project == "unknown project" {
		return "Goal: " + first + " Latest: " + latest
	}
	return "Project: " + project + ". Goal: " + first + " Latest: " + latest
}

func linesFromText(text string, limit int) []string {
	var out []string
	for _, line := range strings.Split(text, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		out = append(out, line)
		if limit > 0 && len(out) >= limit {
			break
		}
	}
	return out
}

func inferStatusFromText(text string) string {
	lower := strings.ToLower(text)
	for _, needle := range []string{"permission", "approve", "waiting for", "needs your input", "blocked"} {
		if strings.Contains(lower, needle) {
			return core.StatusNeedsAttention
		}
	}
	for _, needle := range []string{"completed", "done", "finished"} {
		if strings.Contains(lower, needle) {
			return core.StatusDone
		}
	}
	return ""
}

func confidence(candidates []string) float64 {
	switch {
	case len(candidates) >= 2:
		return 0.6
	case len(candidates) == 1:
		return 0.4
	default:
		return 0.2
	}
}

package core

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

const (
	StatusNeedsAttention = "needs_attention"
	StatusWorking        = "working"
	StatusIdle           = "idle"
	StatusDone           = "done"
	StatusStale          = "stale"
	StatusUnknown        = "unknown"
)

type Session struct {
	ID             string
	Source         string
	NativeID       string
	ProjectPath    string
	Agent          string
	Title          string
	TitleSource    string
	GoalSummary    string
	Status         string
	Pinned         bool
	Tags           string
	LastActivityAt time.Time
	CreatedAt      time.Time
	UpdatedAt      time.Time
	ResumeCommand  string
	AttachCommand  string
	RawPath        string
}

type RuntimeInstance struct {
	ID            string
	SessionID     string
	Backend       string
	NativeID      string
	Surface       string
	ProjectPath   string
	Command       string
	Status        string
	AttachCommand string
	UpdatedAt     time.Time
}

type ScanResult struct {
	Source   string
	Sessions []Session
	Runtimes []RuntimeInstance
	Skipped  bool
	Message  string
	Err      error
}

func CanonicalID(source, nativeID string) string {
	sum := sha1.Sum([]byte(source + "\x00" + nativeID))
	return source + "-" + hex.EncodeToString(sum[:])[:12]
}

func RuntimeID(backend, nativeID string) string {
	sum := sha1.Sum([]byte(backend + "\x00" + nativeID))
	return backend + "-" + hex.EncodeToString(sum[:])[:12]
}

func ProjectID(path string) string {
	if path == "" {
		return ""
	}
	abs, err := filepath.Abs(path)
	if err != nil {
		abs = path
	}
	sum := sha1.Sum([]byte(abs))
	return "project-" + hex.EncodeToString(sum[:])[:12]
}

func ProjectName(path string) string {
	if path == "" {
		return "unknown project"
	}
	clean := filepath.Clean(path)
	base := filepath.Base(clean)
	if base == "." || base == string(filepath.Separator) || base == "" {
		return clean
	}
	return base
}

func DecodeClaudeProjectDir(name string) string {
	if strings.HasPrefix(name, "-") {
		decoded := strings.ReplaceAll(name, "-", string(filepath.Separator))
		if decoded != "" {
			return decoded
		}
	}
	return name
}

func ShellQuote(s string) string {
	if s == "" {
		return "''"
	}
	return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'"
}

func SanitizeName(s string) string {
	s = strings.TrimSpace(strings.ToLower(s))
	re := regexp.MustCompile(`[^a-z0-9_.-]+`)
	s = re.ReplaceAllString(s, "-")
	s = strings.Trim(s, "-")
	if s == "" {
		return "session"
	}
	if len(s) > 48 {
		return strings.Trim(s[:48], "-")
	}
	return s
}

func Truncate(s string, max int) string {
	s = strings.TrimSpace(strings.Join(strings.Fields(s), " "))
	if max <= 0 || len(s) <= max {
		return s
	}
	if max <= 3 {
		return s[:max]
	}
	return strings.TrimSpace(s[:max-3]) + "..."
}

func ExistingDir(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

func ExistingFile(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

func NonZeroTime(t time.Time, fallback time.Time) time.Time {
	if t.IsZero() {
		return fallback
	}
	return t
}

func FormatTime(t time.Time) string {
	if t.IsZero() {
		return "-"
	}
	return t.Local().Format("2006-01-02 15:04")
}

func NewSession(source, nativeID string) Session {
	now := time.Now().UTC()
	return Session{
		ID:        CanonicalID(source, nativeID),
		Source:    source,
		NativeID:  nativeID,
		Status:    StatusUnknown,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

func TmuxAttachCommand(session string) string {
	return fmt.Sprintf("tmux attach -t %s", ShellQuote(session))
}

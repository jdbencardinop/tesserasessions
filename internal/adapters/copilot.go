package adapters

import (
	"context"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/jdbencardinop/tesserasessions/internal/core"
)

type CopilotScanner struct {
	Root string
}

func (s CopilotScanner) Name() string {
	return "copilot"
}

func (s CopilotScanner) Scan(ctx context.Context) core.ScanResult {
	result := core.ScanResult{Source: s.Name()}
	if !core.ExistingDir(s.Root) {
		result.Skipped = true
		result.Message = "Copilot session-state directory not found"
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
		path := filepath.Join(s.Root, entry.Name())
		if strings.HasPrefix(entry.Name(), ".") {
			continue
		}
		info, err := entry.Info()
		if err != nil {
			continue
		}
		if !entry.IsDir() && !isCopilotSessionFile(entry.Name()) {
			continue
		}
		last := info.ModTime()
		if entry.IsDir() {
			last = newestModTime(path, last)
		}
		nativeID := strings.TrimSuffix(entry.Name(), filepath.Ext(entry.Name()))
		session := core.NewSession("copilot", nativeID)
		session.Agent = "copilot"
		session.Title = "Copilot: " + core.Truncate(nativeID, 32)
		session.Status = core.StatusUnknown
		session.LastActivityAt = last.UTC()
		session.CreatedAt = last.UTC()
		session.ResumeCommand = "copilot --resume=" + core.ShellQuote(nativeID)
		session.RawPath = path
		result.Sessions = append(result.Sessions, session)
	}
	sort.Slice(result.Sessions, func(i, j int) bool {
		return result.Sessions[i].LastActivityAt.After(result.Sessions[j].LastActivityAt)
	})
	return result
}

func isCopilotSessionFile(name string) bool {
	switch filepath.Ext(name) {
	case ".json", ".jsonl", ".db", ".md", ".yaml":
		return true
	default:
		return false
	}
}

func newestModTime(root string, fallback time.Time) time.Time {
	newest := fallback
	_ = filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		info, err := d.Info()
		if err == nil && info.ModTime().After(newest) {
			newest = info.ModTime()
		}
		return nil
	})
	return newest
}

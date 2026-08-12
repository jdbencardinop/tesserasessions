package adapters

import (
	"context"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/jdbencardinop/tesserasessions/internal/core"
	"gopkg.in/yaml.v3"
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
		if strings.HasPrefix(entry.Name(), ".") || !entry.IsDir() {
			continue
		}
		path := filepath.Join(s.Root, entry.Name())
		workspacePath := filepath.Join(path, "workspace.yaml")
		if !core.ExistingFile(workspacePath) {
			continue
		}
		info, err := entry.Info()
		if err != nil {
			result.Err = err
			return result
		}
		metadata, err := readCopilotWorkspace(workspacePath)
		if err != nil {
			result.Err = err
			return result
		}
		workspaceModified := info.ModTime()
		if workspaceInfo, statErr := os.Stat(workspacePath); statErr == nil {
			workspaceModified = workspaceInfo.ModTime()
		}
		nativeID := firstNonEmpty(metadata.ID, entry.Name())
		projectPath := firstNonEmpty(metadata.CWD, metadata.GitRoot)
		title := metadata.Name
		if title == "" && metadata.Repository != "" {
			title = "Copilot: " + metadata.Repository
		}
		if title == "" {
			title = "Copilot: " + core.ProjectName(projectPath)
		}
		created := core.NonZeroTime(parseTimeString(metadata.CreatedAt), workspaceModified)
		updated := parseTimeString(metadata.UpdatedAt)
		if updated.IsZero() {
			updated = core.NonZeroTime(created, workspaceModified)
		}
		session := core.NewSession("copilot", nativeID)
		session.ProjectPath = projectPath
		session.Agent = "copilot"
		session.Title = core.Truncate(title, 120)
		session.Status = core.StatusUnknown
		session.LastActivityAt = updated.UTC()
		session.CreatedAt = created.UTC()
		resume := commandWithEnv(
			"COPILOT_HOME",
			filepath.Dir(s.Root),
			"copilot --resume="+core.ShellQuote(nativeID),
		)
		session.ResumeCommand = resumeInDirectory(projectPath, resume)
		session.RawPath = path
		result.Sessions = append(result.Sessions, session)
	}
	sort.Slice(result.Sessions, func(i, j int) bool {
		return result.Sessions[i].LastActivityAt.After(result.Sessions[j].LastActivityAt)
	})
	result.SessionSnapshotComplete = true
	return result
}

type copilotWorkspace struct {
	ID         string `yaml:"id"`
	CWD        string `yaml:"cwd"`
	GitRoot    string `yaml:"git_root"`
	Repository string `yaml:"repository"`
	HostType   string `yaml:"host_type"`
	Branch     string `yaml:"branch"`
	ClientName string `yaml:"client_name"`
	Name       string `yaml:"name"`
	UserNamed  bool   `yaml:"user_named"`
	CreatedAt  string `yaml:"created_at"`
	UpdatedAt  string `yaml:"updated_at"`
}

func readCopilotWorkspace(path string) (copilotWorkspace, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return copilotWorkspace{}, err
	}
	var metadata copilotWorkspace
	if err := yaml.Unmarshal(data, &metadata); err != nil {
		return copilotWorkspace{}, err
	}
	return metadata, nil
}

func parseTimeString(value string) time.Time {
	value = strings.TrimSpace(value)
	for _, layout := range []string{time.RFC3339Nano, time.RFC3339, "2006-01-02 15:04:05"} {
		if parsed, err := time.Parse(layout, value); err == nil {
			return parsed.UTC()
		}
	}
	return time.Time{}
}

package config

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"gopkg.in/yaml.v3"
)

type Config struct {
	DataDir    string        `yaml:"data_dir"`
	Database   string        `yaml:"database"`
	Live       LiveConfig    `yaml:"live"`
	Sources    SourcesConfig `yaml:"sources"`
	Summarizer SummaryConfig `yaml:"summarizer"`
}

type LiveConfig struct {
	DefaultBackend string `yaml:"default_backend"`
}

type SourcesConfig struct {
	ClaudeProjects      string `yaml:"claude_projects"`
	CopilotSessionState string `yaml:"copilot_session_state"`
	HermesDatabase      string `yaml:"hermes_database"`
	OpenCodeDatabase    string `yaml:"opencode_database"`
	CodexHome           string `yaml:"codex_home"`
}

type SummaryConfig struct {
	RemoteCommand string `yaml:"remote_command"`
}

func ConfigPath() string {
	if path := os.Getenv("TSS_CONFIG"); path != "" {
		return path
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "tesserasessions", "config.yaml")
}

func DefaultDataDir() string {
	if path := os.Getenv("TSS_DATA_DIR"); path != "" {
		return path
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".local", "share", "tesserasessions")
}

func DefaultConfig() Config {
	home, _ := os.UserHomeDir()
	claudeRoot := filepath.Join(home, ".claude")
	if env := os.Getenv("CLAUDE_CONFIG_DIR"); env != "" {
		claudeRoot = env
	}
	copilotRoot := filepath.Join(home, ".copilot")
	if env := os.Getenv("COPILOT_HOME"); env != "" {
		copilotRoot = env
	}
	dataDir := DefaultDataDir()
	return Config{
		DataDir:  dataDir,
		Database: filepath.Join(dataDir, "sessions.db"),
		Live: LiveConfig{
			DefaultBackend: "herdr",
		},
		Sources: SourcesConfig{
			ClaudeProjects:      filepath.Join(claudeRoot, "projects"),
			CopilotSessionState: filepath.Join(copilotRoot, "session-state"),
			HermesDatabase:      defaultHermesDatabase(home),
			OpenCodeDatabase:    defaultOpenCodeDatabase(home),
			CodexHome:           defaultCodexHome(home),
		},
	}
}

func Load(path string) (Config, error) {
	cfg := DefaultConfig()
	if path == "" {
		path = ConfigPath()
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return cfg, nil
		}
		return cfg, err
	}
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return cfg, err
	}
	cfg.DataDir = ExpandPath(firstNonEmpty(cfg.DataDir, DefaultDataDir()))
	cfg.Database = ExpandPath(firstNonEmpty(cfg.Database, filepath.Join(cfg.DataDir, "sessions.db")))
	cfg.Sources.ClaudeProjects = ExpandPath(firstNonEmpty(cfg.Sources.ClaudeProjects, DefaultConfig().Sources.ClaudeProjects))
	cfg.Sources.CopilotSessionState = ExpandPath(firstNonEmpty(cfg.Sources.CopilotSessionState, DefaultConfig().Sources.CopilotSessionState))
	cfg.Sources.HermesDatabase = ExpandPath(firstNonEmpty(cfg.Sources.HermesDatabase, DefaultConfig().Sources.HermesDatabase))
	cfg.Sources.OpenCodeDatabase = ExpandPath(firstNonEmpty(cfg.Sources.OpenCodeDatabase, DefaultConfig().Sources.OpenCodeDatabase))
	cfg.Sources.CodexHome = ExpandPath(firstNonEmpty(cfg.Sources.CodexHome, DefaultConfig().Sources.CodexHome))
	if cfg.Live.DefaultBackend == "" {
		cfg.Live.DefaultBackend = "herdr"
	}
	return cfg, nil
}

func EnsureDirs(cfg Config) error {
	return os.MkdirAll(filepath.Dir(cfg.Database), 0755)
}

func ExpandPath(path string) string {
	if path == "" {
		return ""
	}
	if path == "~" {
		home, _ := os.UserHomeDir()
		return home
	}
	if strings.HasPrefix(path, "~/") {
		home, _ := os.UserHomeDir()
		return filepath.Join(home, path[2:])
	}
	return os.ExpandEnv(path)
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}

func defaultHermesDatabase(home string) string {
	hermesHome := strings.TrimSpace(os.Getenv("HERMES_HOME"))
	var base string
	if hermesHome != "" {
		base = ExpandPath(hermesHome)
		if filepath.Base(filepath.Dir(base)) == "profiles" {
			return filepath.Join(base, "state.db")
		}
	} else {
		base = hermesBaseDir(home, runtime.GOOS, os.Getenv("LOCALAPPDATA"))
	}
	if os.Getenv("HERMES_S6_SUPERVISED_CHILD") == "" {
		profile, err := os.ReadFile(filepath.Join(base, "active_profile"))
		if err == nil {
			name := strings.TrimSpace(string(profile))
			if name != "" && name != "default" && filepath.Base(name) == name {
				return filepath.Join(base, "profiles", name, "state.db")
			}
		}
	}
	return filepath.Join(base, "state.db")
}

func defaultOpenCodeDatabase(home string) string {
	dataDir := openCodeDataDir(home)
	if override := os.Getenv("OPENCODE_DB"); override != "" {
		if override == ":memory:" || filepath.IsAbs(override) {
			return override
		}
		return filepath.Join(dataDir, "opencode", override)
	}
	return filepath.Join(dataDir, "opencode", "opencode.db")
}

func openCodeDataDir(home string) string {
	if dataDir := os.Getenv("XDG_DATA_HOME"); dataDir != "" {
		return ExpandPath(dataDir)
	}
	return filepath.Join(home, ".local", "share")
}

func hermesBaseDir(home, goos, localAppData string) string {
	if goos == "windows" {
		if localAppData == "" {
			localAppData = filepath.Join(home, "AppData", "Local")
		}
		return filepath.Join(localAppData, "hermes")
	}
	return filepath.Join(home, ".hermes")
}

func defaultCodexHome(home string) string {
	if codexHome := strings.TrimSpace(os.Getenv("CODEX_HOME")); codexHome != "" {
		return ExpandPath(codexHome)
	}
	return filepath.Join(home, ".codex")
}

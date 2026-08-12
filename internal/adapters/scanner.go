package adapters

import (
	"context"

	"github.com/jdbencardinop/tesserasessions/internal/config"
	"github.com/jdbencardinop/tesserasessions/internal/core"
)

type Scanner interface {
	Name() string
	Scan(context.Context) core.ScanResult
}

func DefaultScanners(cfg config.Config) []Scanner {
	scanners := []Scanner{
		ClaudeScanner{Root: cfg.Sources.ClaudeProjects},
		CopilotScanner{Root: cfg.Sources.CopilotSessionState},
		HermesHistoryScanner{Database: cfg.Sources.HermesDatabase},
		OpenCodeScanner{Database: cfg.Sources.OpenCodeDatabase},
	}
	return append(scanners, LiveScanners()...)
}

func LiveScanners() []Scanner {
	return []Scanner{
		HerdrScanner{},
		TmuxScanner{},
	}
}

func resumeInDirectory(directory, command string) string {
	if directory == "" {
		return command
	}
	return "cd " + core.ShellQuote(directory) + " && " + command
}

func commandWithEnv(name, value, command string) string {
	if value == "" {
		return command
	}
	return name + "=" + core.ShellQuote(value) + " " + command
}

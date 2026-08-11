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
	}
	return append(scanners, LiveScanners()...)
}

func LiveScanners() []Scanner {
	return []Scanner{
		HerdrScanner{},
		TmuxScanner{},
	}
}

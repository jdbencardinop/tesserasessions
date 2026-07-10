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
	return []Scanner{
		ClaudeScanner{Root: cfg.Sources.ClaudeProjects},
		CopilotScanner{Root: cfg.Sources.CopilotSessionState},
		HerdrScanner{},
		TmuxScanner{},
	}
}

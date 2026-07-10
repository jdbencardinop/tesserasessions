# Exploration: cli-inventory-core

## Relevant files

- `cmd/tss/main.go`: CLI entrypoint.
- `internal/cli/root.go`: command surface.
- `internal/config/config.go`: default paths and environment overrides.
- `internal/core/model.go`: normalized session/runtime models and helpers.
- `internal/store/store.go`: SQLite schema, migrations, upserts, queries.
- `internal/adapters/scanner.go`: scanner interface and default scanner list.
- `internal/adapters/claude.go`: Claude Code JSONL metadata scanning.
- `internal/adapters/copilot.go`: Copilot session-state metadata scanning.
- `internal/adapters/herdr.go`: optional Herdr live agent discovery.
- `internal/adapters/tmux.go`: optional tmux pane discovery.
- `internal/adapters/text.go`: lazy text candidate extraction.
- `internal/summarize/summarize.go`: local/extractive summaries.
- `internal/**/*_test.go`: current validation coverage.

## Minimal changeset

This is a backfill record for the initial source tree. The intended foundation scope is `go.mod`, `go.sum`, `cmd/`, and `internal/`.

## Validation

- `go test ./...`
- `tss scan`
- `tss list`
- disposable tmux scan/show/send command-generation check


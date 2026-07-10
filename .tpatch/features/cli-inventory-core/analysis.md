# Analysis: cli-inventory-core

## Summary

Backfill the foundation feature for the already-created `tss` CLI. This feature covers the initial standalone Go/Cobra CLI, local SQLite inventory, Claude and Copilot metadata scanners, Herdr/tmux live-runtime scaffolding, and baseline commands.

## Compatibility

- Compatible with the selected standalone Go CLI architecture.
- Uses local-only SQLite via `modernc.org/sqlite`.
- Scanners are read-only against external agent stores.
- Herdr and tmux integrations are optional and skip cleanly when unavailable.

## Affected areas

- `go.mod`, `go.sum`
- `cmd/tss/main.go`
- `internal/cli/root.go`
- `internal/config/config.go`
- `internal/core/model.go`
- `internal/store/store.go`
- `internal/adapters/*.go`
- `internal/summarize/summarize.go`
- package tests under `internal/**`

## Risks

- This feature is being backfilled after implementation, so the recorded patch includes some state that later slices may also touch.
- The repo was initialized after code existed, so the patch is a broad initial-addition snapshot rather than a clean chronological slice.
- Future tpatch workflows should register and advance feature metadata before implementation to avoid this overlap.


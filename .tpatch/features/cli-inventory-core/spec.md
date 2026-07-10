# Spec: cli-inventory-core

## Problem

Create a standalone local-first CLI that inventories coding-agent sessions across local stores and live terminal backends.

## Acceptance criteria

1. A Go module builds a `tss` CLI from `cmd/tss`.
2. `tss doctor` reports config, source paths, optional backends, and inventory count.
3. `tss scan` indexes Claude and Copilot session metadata into SQLite.
4. `tss list` and `tss show` expose normalized session metadata.
5. Herdr and tmux adapters skip cleanly when missing or not running.
6. Attach/open/send command scaffolding exists with `--print` for review.
7. Local/extractive summary command exists.
8. `go test ./...` passes.

## Out of scope

- Full GUI/TUI.
- Deep Hermes/t3code adapters.
- Hosted sync or telemetry.
- Remote LLM summarization by default.

## Plan

1. Initialize Go module and dependencies.
2. Add core model/config/store packages.
3. Add Claude and Copilot historical scanners.
4. Add Herdr and tmux live scanner scaffolding.
5. Add Cobra commands for scan/list/show/doctor/summarize/attach/open/send.
6. Add initial tests for core IDs, quoting, store upserts, and Claude scanning.


# Spec: docs-guide-cheatsheet

## Problem

The CLI has working inventory commands but no user-facing guide. Users need a quick way to understand what `tss` does, how to run it, what it stores, how live attach/open/send behavior works, and how feature work is tracked with tpatch.

## Acceptance criteria

1. `README.md` explains scope, quick start, command overview, architecture, data/config locations, privacy defaults, and tpatch workflow.
2. `docs/cheatsheet.md` provides concise command examples for scanning, listing, showing, summarizing, attaching, opening, sending, config overrides, source adapters, and development checks.
3. `docs/roadmap.md` maps the plan to registered tpatch feature slugs and documents the dependency DAG.
4. Documentation accurately reflects the current CLI help output from `go run ./cmd/tss --help`.
5. `go test ./...` passes after the docs/tpatch changes.

## Out of scope

- Publishing hosted docs.
- Adding a GUI.
- Implementing new CLI behavior beyond documentation and tpatch registration.
- Completing full tpatch `record`/`land` flows before this directory is a git repository.

## Plan

1. Initialize tpatch in the repo.
2. Register roadmap features with explicit slugs.
3. Add dependency edges with `cli-inventory-core` as the foundation.
4. Set `test_command` to `go test ./...`.
5. Add README, cheatsheet, and roadmap docs.
6. Validate tpatch DAG and Go tests.


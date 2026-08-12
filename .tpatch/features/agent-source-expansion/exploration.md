# Exploration: agent-source-expansion

## Existing files

- `internal/adapters/claude.go`
  - `ClaudeScanner.Scan`;
  - `readJSONLMetadata`;
  - `readTextCandidatesFromJSONL`.
- `internal/adapters/copilot.go`
  - `CopilotScanner.Scan`;
  - current modification-time heuristics.
- `internal/adapters/scanner.go`
  - `DefaultScanners`.
- `internal/config/config.go`
  - `SourcesConfig`, defaults, and path expansion.
- `internal/cli/root.go`
  - `scanCmd`, `doctorCmd`, and `sourcePath`.
- `internal/core/model.go`
  - normalized `Session` fields and shell quoting.
- `internal/adapters/claude_test.go`
  - current synthetic Claude fixture.

## Proposed files

- `internal/adapters/sqlite.go`
  - stability-checked private DB+WAL/rollback-journal snapshot, private
    recovery, and `query_only=1` metadata access.
- `internal/adapters/hermes.go`
  - Hermes v25 metadata query and timestamp mapping.
- `internal/adapters/hermes_test.go`
  - synthetic current-schema compression fixture and read-only guarantees.
- `internal/adapters/opencode.go`
  - current `session`/`project` join.
- `internal/adapters/opencode_test.go`
  - synthetic current-schema/integrity fixture and read-only guarantees.
- `internal/adapters/copilot_test.go`
  - YAML fixtures for named and unnamed sessions.
- `docs/source-adapters.md`
  - supported stores, discovery overrides, privacy boundary, and deferred
    sources.

## Public schema references

### Hermes Agent

- `hermes_constants.py` at
  `96b922732213109c020f0273a37788df6e7d7f9c`;
- `hermes_state_common.py` at
  `e386c1260b3fe741b29493f5b46a8a5053ab0871`;
- `hermes_state.py` at
  `2bb6c58fd417fcde97e338ad61db3459d3e37783`;
- `hermes_cli/sessions_cmd.py` at
  `095ad5f9e7e2a410858d57a8a64eaec4ed6f1f88`.

### OpenCode

- `packages/core/src/database/database.ts` at `d61adf04`;
- `packages/core/src/global.ts` at `a192a4b4`;
- `packages/core/src/session/sql.ts` at `264a1d2c`;
- `packages/core/src/project/sql.ts` at `ab05fdac`;
- `packages/opencode/src/cli/cmd/tui.ts` at `95ffac7e`;
- `packages/opencode/src/cli/cmd/run.ts` at `3927f615`.

## Minimal change set

- Extend source configuration with Hermes/OpenCode database paths.
- Add two SQLite scanners and tests.
- Refine, rather than duplicate, Claude/Copilot scanners.
- Add authoritative historical-session reconciliation while preserving manual
  inventory metadata.
- Update scanner registration, doctor output, README, cheatsheet, roadmap,
  changelog, and a focused source-adapter guide.

## Validation

```sh
go test ./internal/adapters ./internal/config ./internal/cli -count=1
make test
make lint
make build
go test -race ./internal/adapters ./internal/cli -count=1
git diff --check
tpatch feature deps --validate-all
```

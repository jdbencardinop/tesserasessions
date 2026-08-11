# Exploration: runtime-status-contract

## Relevant files and symbols

- `internal/cli/root.go`
  - `Execute` registers commands.
  - `scanCmd` currently combines adapter discovery with database upserts.
  - Existing JSON output writes database models directly.
- `internal/adapters/scanner.go`
  - `Scanner` and `DefaultScanners` are the current adapter seam.
- `internal/adapters/herdr.go`
  - Produces semantic Herdr runtime state and verified foreground cwd when
    available.
- `internal/adapters/tmux.go`
  - Produces tmux pane/cwd observations but weak semantic state.
- `internal/core/model.go`
  - `Session`, `RuntimeInstance`, and the current mixed status constants.
- `internal/store/store.go`
  - `UpsertRuntime` only inserts/updates.
  - `RuntimeForSession` queries by `session_id`.
  - No runtime-generation pruning exists.
- `internal/config/config.go`
  - Existing provider paths and live backend preference.
- `internal/adapters/herdr_test.go` and `internal/cli/live_commands_test.go`
  - Existing live-adapter and command-generation test patterns.

## Proposed implementation shape

- Add focused wire DTOs under the existing core/CLI packages rather than
  exposing database structs.
- Add a dedicated Cobra status command that reads `cmd.InOrStdin()` and writes
  `cmd.OutOrStdout()` for testability.
- Separate live observation collection from persistence so status probes can
  reuse adapters without calling `scanCmd`.
- Add matcher/aggregator helpers in a focused internal package or file with no
  CLI dependencies.
- Add scan-generation reconciliation to the existing store package.

## Minimal first slice

1. Schema v1 DTOs and stdin/stdout command plumbing.
2. Herdr and tmux in-memory observations with bounded contexts.
3. Path/branch matching and raw plus aggregate responses.
4. Provider availability and per-query errors.
5. Tests proving the command does not mutate the database.

Generic external direct-process observation and additional agent sources remain
follow-up work.

## Validation

```sh
go test ./...
make lint
git diff --check
tpatch feature deps --validate-all
```


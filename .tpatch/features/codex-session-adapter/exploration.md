# Exploration: codex-session-adapter

## Existing integration points

- `internal/adapters/scanner.go`
  - `DefaultScanners`;
  - source-qualified resume helpers.
- `internal/adapters/claude.go`
  - bounded JSONL line-reading precedent.
- `internal/config/config.go`
  - `SourcesConfig`, defaults, and environment overrides.
- `internal/cli/root.go`
  - scan source help, doctor paths, and `sourcePath`.
- `internal/core/model.go`
  - normalized session shape and authoritative snapshot marker.
- `internal/store/store.go`
  - `ReplaceSessions` authoritative reconciliation.
- `docs/source-adapters.md`
  - public source contract and privacy boundary.

## Proposed files

- `internal/adapters/codex.go`
  - rollout discovery, first-line parsing, source filtering, names, deduplication.
- `internal/adapters/codex_test.go`
  - synthetic active/legacy/filter/dedupe/error fixtures.
- `internal/config/config_test.go`
  - `CODEX_HOME` default/override coverage.

## Upstream references

Pinned research commit:
`4ef836f883c38ba6d39e6920f335ce6452b7de33`.

- `codex-rs/utils/home-dir/src/lib.rs`
  - `CODEX_HOME` resolution.
- `codex-rs/rollout/src/lib.rs`
  - `SESSIONS_SUBDIR`, `ARCHIVED_SESSIONS_SUBDIR`, interactive sources.
- `codex-rs/rollout/src/rollout_file_name.rs`
  - rollout filename/revert conventions.
- `codex-rs/history/src/lib.rs`
  - flattened `RolloutLine`.
- `codex-rs/history/src/rollout_payload.rs`
  - `#[serde(tag="type", rename_all="snake_case")]`.
- `codex-rs/protocol/src/protocol.rs`
  - `SessionMeta`, `SessionMetaLine`, `SessionSource`.
- `codex-rs/rollout/src/session_index.rs`
  - `SessionIndexEntry`.
- `codex-rs/cli/src/main.rs`
  - top-level `codex resume`.
- `codex-rs/rollout/src/list.rs`
  - active directory traversal and thread metadata semantics.

## Minimal change set

- Add one JSONL scanner and tests.
- Extend source configuration and CLI discovery.
- Reuse `ReplaceSessions`; no store migration is required.
- Update existing source documentation/roadmap only.
- Add no third-party compression dependency.

## Validation

```sh
go test ./internal/adapters ./internal/config ./internal/cli -count=1
make test
make lint
make build
go test -race ./internal/adapters ./internal/config ./internal/cli -count=1
git diff --check
tpatch feature deps --validate-all
```


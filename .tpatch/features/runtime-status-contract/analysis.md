# Analysis: runtime-status-contract

## Summary

Add a stable, read-only status-provider contract so workspace orchestrators can
enrich their own views with live agent runtime state. `tws` remains authoritative
for workspace topology, Git state, locks, failed stages, and sessions it
launches. `tss` reports semantic agent state and externally observed runtimes.

The first consumer is the `tesseraworkspaces` `agent-work-status-dashboard`
feature. Its baseline dashboard shipped independently in `tws` v1.2.12.

## Decisions

- Use a versioned CLI JSON contract, not SQLite access, Go imports, or a plugin.
- The command is side-effect-free: no implicit `tss scan` and no database writes.
- Accept a batch request envelope on stdin and emit one response envelope.
- Echo an opaque `query_id`; do not treat path as identity.
- Accept verified path plus optional repository and Git branch hints.
- Return raw matched runtimes and separate aggregate axes:
  - `runtime_presence`: `present`, `absent`, `stale`, or `unknown`;
  - `agent_state`: `working`, `ready`, `blocked`, `done`, or `unknown`.
- Never guess a match for a historical session without a verified project path.
- `tws` retains final rollup authority and its own local signals take precedence.

## Current compatibility gaps

- `core.Session` and `core.RuntimeInstance` expose raw Go JSON field names.
- The existing status enum mixes runtime presence with semantic agent state.
- Runtime rows are only upserted; disappeared runtimes are not reconciled.
- Runtime lookup is by `tss` session ID, not a consumer query key or path.
- Copilot historical sessions currently have no verified project path.
- `tss` has no generic direct-process observer. Consumers must receive explicit
  coverage/availability metadata rather than an incorrect `absent`.

## Matching constraints

Paths alone cannot distinguish checkout branches sharing one physical
repository. Matching therefore needs:

1. a required opaque `query_id`;
2. an optional verified absolute path;
3. optional `repo_root` and `git_branch` hints;
4. canonical path normalization;
5. exact or ancestor matching with deepest-path-wins;
6. explicit match evidence in the response.

## Risks

- A provider timeout must not erase a consumer's baseline result.
- A cached row must not be reported as live without freshness evidence.
- Multiple runtimes can map to one query and need deterministic aggregation.
- Symlinks, case-insensitive filesystems, moved paths, and nested working
  directories can otherwise produce false misses or false matches.

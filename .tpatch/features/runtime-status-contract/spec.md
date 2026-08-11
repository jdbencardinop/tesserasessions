# Spec: runtime-status-contract

## Problem

External orchestrators need a stable way to query live agent state without
depending on `tss` storage internals or transferring final status ownership to
`tss`.

## Command contract

The planned command is:

```sh
tss status --json < request.json
```

It performs bounded live probes in memory, may read cached observations only as
explicitly stale evidence, and does not scan or mutate the inventory database.

## Acceptance criteria

1. Register `runtime-status-contract` in tpatch with hard dependencies on
   `cli-inventory-core` and `live-herdr-tmux-control`.
2. `tss status --json` accepts one versioned batch request JSON object on stdin.
3. Every query has a required `query_id` and may include `path`, `repo_root`,
   `git_branch`, and consumer metadata that `tss` echoes but does not interpret.
4. The command is side-effect-free:
   - no implicit `tss scan`;
   - no database writes;
   - no agent input, attach, focus, or process mutation.
5. The response includes:
   - `schema_version`;
   - provider-level `observed_at`, `fresh_until`, availability, and errors;
   - one result per input `query_id`;
   - per-query errors without failing unrelated queries;
   - match type and matched path;
   - raw matched runtimes;
   - aggregate `runtime_presence` and `agent_state`.
6. Raw runtime observations include a stable observation ID, backend, native
   target, observed path, optional repository/branch hints, semantic state,
   liveness/freshness, `observed_at`, and `expires_at`.
7. Matching is deterministic:
   - resolve absolute paths and symlinks where possible;
   - normalize separators, trailing slashes, and case on case-insensitive hosts;
   - derive `repo_root` from Git's common directory so linked worktrees share
     the main repository identity;
   - support exact and ancestor path matches;
   - choose the deepest matching query path;
   - use `repo_root` plus `git_branch` to disambiguate equally specific/shared
     paths when runtime evidence contains those fields;
   - reject known repository or branch conflicts instead of falling back to
     path-only matching;
   - leave cwd-less or unverifiable sessions unmatched.
8. Multiple matches are reduced deterministically:
   - presence: live `present` beats `stale`, which beats `unknown`; `absent`
     requires completed provider coverage with no match;
   - agent state: `blocked` beats `working`, then `ready`, `done`, and `unknown`;
   - raw matches remain available so consumers can apply a different rollup.
9. Provider failures are bounded and explicit. A missing, timed-out, or
   incompatible adapter returns availability/error metadata rather than a false
   `absent`. Metadata enrichment timeouts also make coverage incomplete.
10. Runtime persistence is reconciled by explicit scans:
    - a successful adapter scan removes or expires rows absent from that scan
      generation;
    - a failed adapter scan does not delete its prior rows;
    - persisted rows never establish live `present` after their expiry.
11. JSON DTOs use explicit lowercase `json` tags and a schema version; the
    wire contract is not the raw database model.
12. Tests cover malformed envelopes, mixed per-query success/error, shared
    repository paths with branch hints, nested cwd matching, multiple-runtime
    aggregation, metadata-enrichment timeouts, linked Git worktrees, current
    Herdr response shape, tmux shell filtering, stale observations, and
    unchanged database state.

## Consumer precedence

For the planned `tws` consumer:

```text
needs_attention: tws local failure/stale state OR tss blocked/stale
active:          tws-owned live session OR tss runtime present
idle:            neither condition above
```

`tws` is authoritative for sessions it launches. `tss` enrichment must never
override a live direct/tmux session known to `tws`. Missing or incompatible
`tss` only changes provider availability metadata.

## Out of scope

- Implementing `tws status`.
- Implementing the later `tws` `tss-status-enrichment` adapter.
- Reading the `tss` SQLite database from another process.
- A Go library dependency between the repositories.
- A plugin system or long-running daemon.
- Guessing paths for historical Copilot or other cwd-less sessions.
- Generic external direct-process discovery in the first contract slice.

## Phased implementation

1. Add versioned request/response DTOs and validation.
2. Add side-effect-free live observation collection.
3. Add matching, aggregation, freshness, and per-query errors.
4. Reconcile persisted runtime rows during explicit `scan`.
5. Publish examples and consumer compatibility tests.

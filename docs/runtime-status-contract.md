# Runtime status provider contract

This document coordinates `tesserasessions` (`tss`) with consumers such as
`tesseraworkspaces` (`tws`).

## Ownership

`tws` owns:

- workspace, feature, branch, and worktree topology;
- sessions that `tws` launches;
- direct-process and tmux liveness for those sessions;
- locks, stale state, failed stages, and final dashboard rollup.

`tss` owns:

- cross-agent runtime observation;
- semantic agent state such as working, ready, blocked, or done;
- externally launched runtimes visible through Herdr, tmux, or later adapters.

The integration is an optional CLI JSON contract. Consumers must not read the
`tss` SQLite database or import `tss` Go packages.

## Rollout

The work proceeds in parallel:

1. `tws/agent-work-status-dashboard`
   - shipped in `tws` v1.2.12;
   - reports topology, runtime presence, attention rollups, direct/tmux
     sessions, sync failures, and stale state without `tss` coupling;
   - external direct sessions without durable evidence remain `unknown`.
2. `tss/runtime-status-contract`
   - implemented and verified;
   - exposes `tss status --json` with bounded live probes and schema v1.
3. `tws/tss-status-enrichment`
   - later optional child feature;
   - invoke `tss` with a timeout and schema guard;
   - retain the baseline result when `tss` is unavailable.

## Planned request

```json
{
  "schema_version": 1,
  "queries": [
    {
      "query_id": "auth/models",
      "path": "/workspaces/auth/models",
      "repo_root": "/repos/app",
      "git_branch": "feature/auth-models",
      "metadata": {
        "workspace_id": "ws-123"
      }
    }
  ]
}
```

`query_id` is the consumer's opaque correlation key. Path is evidence, not
identity. Git hints distinguish checkout branches that share one physical
repository path.

## Planned response

```json
{
  "schema_version": 1,
  "observed_at": "2026-08-11T08:00:00Z",
  "fresh_until": "2026-08-11T08:00:10Z",
  "providers": [
    {
      "name": "herdr",
      "available": true,
      "observed_at": "2026-08-11T08:00:00Z",
      "fresh_until": "2026-08-11T08:00:10Z"
    },
    {
      "name": "tmux",
      "available": true,
      "observed_at": "2026-08-11T08:00:00Z",
      "fresh_until": "2026-08-11T08:00:10Z"
    }
  ],
  "results": [
    {
      "query_id": "auth/models",
      "metadata": {
        "workspace_id": "ws-123"
      },
      "runtime_presence": "present",
      "agent_state": "blocked",
      "match": {
        "type": "repo_branch",
        "matched_path": "/workspaces/auth/models"
      },
      "runtimes": [
        {
          "observation_id": "herdr:w1:p2",
          "backend": "herdr",
          "native_id": "w1:p2",
          "path": "/workspaces/auth/models",
          "repo_root": "/repos/app",
          "git_branch": "feature/auth-models",
          "runtime_presence": "present",
          "agent_state": "blocked",
          "observed_at": "2026-08-11T08:00:00Z",
          "expires_at": "2026-08-11T08:00:10Z"
        }
      ],
      "errors": []
    }
  ]
}
```

Schema v1 is implemented by `tss status --json`. Additive fields may be
introduced compatibly; breaking field or enum changes require a new schema
version.

## Side effects and freshness

The status command:

- does not run `tss scan`;
- does not write the inventory database;
- does not send input or focus/attach to an agent;
- performs bounded live probes in memory;
- returns timestamps, expiry, provider coverage, and per-query errors.

Explicit `tss scan` remains responsible for persisted inventory. Successful
runtime scans must reconcile disappeared rows; failed provider scans must not
erase prior evidence.

## Matching

- Canonicalize absolute paths and resolve symlinks where possible.
- Treat nested foreground working directories as descendant matches.
- Assign a runtime to the deepest matching query path.
- Use repository and branch hints when physical paths are shared.
- Return match evidence.
- Leave cwd-less or unverifiable sessions unmatched.

## Consumer rollup

`tws` keeps final authority:

```text
needs_attention: local failure/stale state OR tss blocked/stale
active:          tws-owned live session OR tss runtime present
idle:            neither condition above
```

`runtime_presence` and `agent_state` remain separate. In particular, `ready`
means a live agent can accept input; it is not the dashboard's no-session
`idle`.

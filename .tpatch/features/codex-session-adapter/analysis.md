# Analysis: codex-session-adapter

## Summary

Add a separate, read-only OpenAI Codex CLI historical-session adapter after the
four-source Claude/Copilot/Hermes/OpenCode feature.

Current Codex (research ref
`4ef836f883c38ba6d39e6920f335ce6452b7de33`) uses authoritative JSONL rollout
files under `${CODEX_HOME}/sessions`, defaulting to `~/.codex/sessions`.
`state.db` is a rebuildable metadata index and is not needed for a correct first
adapter. Archived rollouts are moved under `archived_sessions`.

## Current persistence contract

### Layout

```text
${CODEX_HOME}/
├── sessions/YYYY/MM/DD/rollout-<timestamp>-<thread-id>.jsonl
├── sessions/rollout-...jsonl
├── archived_sessions/...
├── session_index.jsonl
└── state.db
```

The flat `sessions/*.jsonl` level remains relevant for legacy rollouts.
Revert variants may append a rollout UUID after the stable thread UUID, so file
names are not the canonical identity.

### First-line metadata

Each rollout starts with a flattened `RolloutLine`:

```json
{
  "timestamp": "2026-01-27T12:34:56Z",
  "ordinal": 0,
  "type": "session_meta",
  "payload": {
    "session_id": "...",
    "id": "<stable-thread-id>",
    "parent_thread_id": null,
    "forked_from_id": null,
    "timestamp": "2026-01-27T12:34:56Z",
    "cwd": "/absolute/project",
    "source": "cli",
    "history_mode": "legacy",
    "git": {
      "branch": "main"
    }
  }
}
```

`id` is the stable thread identity and exact resume key. `session_id` groups
root and subagent threads and must not replace the thread ID.

### Independently resumable sessions

Codex's interactive source set is:

- `cli`;
- `vscode`;
- custom `atlas`;
- custom `chatgpt`.

`exec`, `mcp`, internal, unknown, and subagent sources are not top-level
interactive sessions. `parent_thread_id != null` also identifies subagents.
Forks with no parent remain independently resumable.

### Names and recency

There is no durable title in rollout metadata. `session_index.jsonl` is an
append-only name-to-thread index:

```json
{"id":"<thread-id>","thread_name":"refactor-auth","updated_at":"..."}
```

The last entry for an ID is its current name. It is optional enrichment, not
enumeration authority. Creation time comes from `payload.timestamp` (outer
timestamp fallback); recency uses rollout file modification time.

### Resume

Interactive exact resume is:

```sh
CODEX_HOME='<home>' codex resume '<thread-id>'
```

## Local availability

No Codex store is installed on this machine. Public-schema research and
synthetic fixtures therefore provide correctness evidence. The source must skip
cleanly when `${CODEX_HOME}/sessions` is absent.

## Risks and boundaries

- Read only the first rollout line. Do not read later prompts, responses, or
  tool calls, and do not extract or retain embedded base instructions, dynamic
  tools, or repository URLs from the metadata object.
- Report malformed/non-session first lines as an incomplete snapshot rather
  than treating a populated store as empty. Refresh valid rows, but suppress
  authoritative pruning while any rollout is unreadable or concurrently
  partial.
- Deduplicate multiple rollouts for the same thread, choosing the newest file
  deterministically.
- A failed/incomplete traversal must not prune prior inventory rows.
- `.jsonl.zst`, archived sessions, SQLite acceleration, and preview extraction
  are deferred to avoid a new compression dependency and transcript-content
  reads.

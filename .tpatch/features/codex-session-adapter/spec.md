# Spec: codex-session-adapter

## Problem

Codex CLI sessions are absent from `tss`. Users need independently resumable
interactive Codex threads in the same metadata-first inventory as Claude,
Copilot, Hermes, and OpenCode.

## Acceptance criteria

1. Use tpatch Path B only; author artifacts directly and never invoke the tpatch
   AI provider.
2. Register `codex-session-adapter` with a hard dependency on
   `agent-source-expansion`.
3. Add `sources.codex_home` configuration:
   - `CODEX_HOME` overrides the path;
   - default is `~/.codex`;
   - generated resume commands preserve the resolved `CODEX_HOME`.
4. Enumerate active rollout files under `${CODEX_HOME}/sessions`:
   - nested date layout;
   - legacy flat layout;
   - `.jsonl` only in this slice;
   - never traverse `archived_sessions`.
5. Read only a bounded first JSONL line and require:
   - outer `type == "session_meta"`;
   - payload object;
   - non-empty stable `payload.id`;
   - non-empty absolute `payload.cwd`;
   - valid creation timestamp;
   - supported history mode (`legacy` or `paginated`).
6. Include only independently resumable interactive sources:
   - `cli`, `vscode`, custom `atlas`, custom `chatgpt`;
   - exclude rows with `parent_thread_id`;
   - exclude exec/MCP/internal/subagent/unknown sources.
7. Use `payload.id` as `native_id`, never filename or shared `session_id`.
8. Parse optional names from `${CODEX_HOME}/session_index.jsonl`:
   - last valid entry per ID wins;
   - malformed index rows are ignored because the index is auxiliary;
   - fallback title is `Codex: <project>`.
9. Use rollout file modification time as `last_activity_at` and metadata
   timestamp as `created_at`.
10. Emit exact resume command:

    ```sh
    cd '<cwd>' && CODEX_HOME='<home>' codex resume '<thread-id>'
    ```

11. Deduplicate duplicate/revert rollouts by thread ID:
    - newest modification time wins;
    - lexical path breaks equal-time ties.
12. Successful scans set `SessionSnapshotComplete` and reconcile removed
    active sessions while preserving manual inventory metadata.
13. Missing sessions directory is a clean skip. Traversal errors are fatal.
    Malformed or concurrently partial rollout metadata produces an explicit
    incomplete-snapshot warning: valid rows still refresh, but prior rows are
    not pruned.
14. Synthetic fixtures cover:
    - nested and flat layouts;
    - source filtering and subagent exclusion;
    - custom source encoding;
    - names and rename precedence;
    - duplicate rollout selection;
    - malformed first line and oversized first line;
    - custom `CODEX_HOME` resume;
    - authoritative reconciliation.
15. Update doctor, scan help, README, cheatsheet, roadmap, changelog, and source
    adapter documentation.
16. Full, lint, build, race, smoke, and diff gates pass with iterative reviewer
    approval.

## Out of scope

- `.jsonl.zst` compressed rollouts.
- Archived session inventory.
- `state.db` acceleration or schema coupling.
- Reading first user message/preview or other transcript content.
- Codex exec/noninteractive sessions.
- t3code sessions.

## Implementation plan

1. Add Codex path configuration and default discovery.
2. Add bounded first-line rollout and session-index parsing.
3. Add source filtering, deduplication, normalized sessions, and exact resume.
4. Wire scanner, doctor, source paths, reconciliation, and docs.
5. Add synthetic fixtures and missing-store smoke checks.
6. Run independent schema/code review loops and land the feature.

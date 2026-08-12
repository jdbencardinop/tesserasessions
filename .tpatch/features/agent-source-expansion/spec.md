# Spec: agent-source-expansion

## Problem

The inventory currently lists Claude and Copilot sessions, but both adapters
discard authoritative metadata. Hermes and OpenCode have no historical adapters.
Users need a consistent, read-only inventory with exact session identity,
project path, title, timestamps, and resume commands.

## Acceptance criteria

1. Preserve tpatch Path B: artifacts are authored directly and phases advance
   with `--manual`; the configured tpatch AI provider is not used.
2. Amend the feature scope to:
   - harden Claude Code;
   - harden GitHub Copilot CLI;
   - add Hermes Agent SQLite;
   - add OpenCode SQLite;
   - defer Codex and t3code.
3. Claude scanning:
   - enumerates only direct-child session JSONL files and excludes subagent
     transcript directories;
   - reads exact `sessionId`, `cwd`, timestamps, Git branch, and title metadata
     from JSONL;
   - prefers `customTitle`, then `aiTitle`, then `agentName`, with a project
     fallback;
   - uses the transcript directory name only when verified JSON metadata is
     absent;
   - emits `claude --resume <id>` in the verified cwd.
4. Copilot scanning:
   - enumerates session directories containing `workspace.yaml`;
   - parses exact ID, cwd/git root, repository, branch, name, created/updated;
   - uses `name`, then repository/project fallback for display title;
   - emits `copilot --resume=<id>` in the verified cwd;
   - does not guess metadata from `session.db` or transcript content.
5. Hermes scanning:
   - discovers `${HERMES_HOME}/state.db`, default/profile paths, and explicit
     config;
   - queries a private, source-stable SQLite snapshot with `query_only=1`;
   - lists one logical unarchived root `source='cli'` session and projects
     compression lineages to the latest continuation metadata;
   - uses exact ID, title, cwd/Git root, started time, and freshest metadata-only
     activity timestamp;
   - emits `hermes --resume <id>` in the verified cwd;
   - never reads message content or gateway PII.
6. OpenCode scanning:
   - honors `OPENCODE_DB`, XDG/platform data paths, and explicit config;
   - queries a private, source-stable SQLite snapshot with `query_only=1`;
   - joins current `session` and `project` tables;
   - rejects orphaned session/project relationships as malformed state;
   - lists non-archived root sessions;
   - uses exact ID, title, directory/worktree, created/updated epoch-ms values;
   - emits `opencode --session <id>` in the verified directory.
7. `doctor`, `scan --source`, docs, and config expose all four historical
   sources.
8. Resume commands preserve the discovered/configured source store by setting
   `CLAUDE_CONFIG_DIR`, `COPILOT_HOME`, `HERMES_HOME`, or `OPENCODE_DB` before
   invoking the native exact-session command.
9. A missing store is a clean skipped provider. A malformed current store is an
   explicit error and does not create or alter files.
10. Successful authoritative historical scans atomically remove sessions that
   were deleted, archived, or filtered out at the source while preserving
   manual titles, pins, and tags on remaining sessions. Failed/missing scans do
   not prune prior rows.
11. Synthetic fixtures cover:
   - Claude exact cwd/title/session resume and hyphenated paths;
   - oversized Claude payload rows followed by valid metadata;
   - Copilot named/unnamed workspace metadata;
   - Hermes compression lineage, active/root filtering, timestamp conversion,
     and read-only open;
   - OpenCode active/root filtering, path fallback, timestamp conversion, and
     read-only open;
   - missing and malformed databases;
   - a live WAL database whose source DB/WAL/SHM bytes remain unchanged.
12. Local smoke scans show Claude/Copilot sessions with verified project paths
    and names where source metadata provides them.
13. `make test`, `make lint`, `make build`, race tests, and `git diff --check`
    pass.

## Out of scope

- Codex historical sessions.
- t3code application/server sessions.
- OpenCode legacy JSON storage and every development-channel DB.
- Hermes gateway routing sessions without a coding-workspace cwd.
- Independently listable Hermes branch conversations.
- Reading raw message/transcript content for titles or summaries.
- Changing the live runtime status contract.

## Implementation plan

1. Add source config and platform-aware default path helpers.
2. Add a shared read-only SQLite opener.
3. Refactor Claude metadata extraction and exact resume behavior.
4. Replace Copilot filesystem heuristics with `workspace.yaml`.
5. Add Hermes and OpenCode scanners with synthetic SQLite fixtures.
6. Wire scanners into default discovery, doctor, and documentation.
7. Run local/synthetic validation and independent reviewer passes.

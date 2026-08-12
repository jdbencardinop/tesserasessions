# Analysis: agent-source-expansion

## Summary

Expand the historical session inventory from two shallow sources to four
reliable, metadata-first sources:

- harden the existing Claude Code adapter;
- harden the existing GitHub Copilot CLI adapter;
- add Hermes Agent's current SQLite store;
- add OpenCode's current SQLite store.

Codex and t3code remain later features. The original roadmap named Hermes first,
deferred t3code until the core stabilized, and listed OpenCode as an optional
later adapter. The runtime status contract is now delivered, so OpenCode can be
included without blocking the cross-tool status work.

## Current state

### Claude Code

The existing scanner enumerates `~/.claude/projects/*/*.jsonl`, but derives the
project path by replacing hyphens in the encoded directory name. That transform
is ambiguous for real paths containing hyphens. Current transcript rows already
contain exact `cwd`, `sessionId`, `timestamp`, and `gitBranch` fields. Metadata
events also expose `customTitle`, `aiTitle`, and `agentName`.

The existing resume command uses `claude -c`, which resumes the latest project
session rather than the indexed session. Claude supports exact
`claude --resume <id>`.

### GitHub Copilot CLI

The existing scanner treats each top-level entry as a session and uses file
modification time, leaving all project paths empty. Current session directories
contain authoritative `workspace.yaml` with:

- `id`;
- `cwd` and optional `git_root`;
- optional `repository`, `branch`, and `name`;
- `created_at` and `updated_at`.

Exact resumption uses `copilot --resume=<id>`.

### Hermes Agent

Current Hermes Agent (research ref
`5fffe560661c87d988c4ef2834df14bfb8acba55`) stores sessions in SQLite schema
v25 at `${HERMES_HOME}/state.db`, normally `~/.hermes/state.db`. Named profiles
use `~/.hermes/profiles/<name>/state.db`.

The `sessions` table contains canonical ID, source, title, cwd, Git root/branch,
creation/activity timestamps, parent lineage, and archive state. A read-only
inventory should list unarchived root CLI sessions and never read raw message
content, origin/gateway identifiers, prompts, or credentials.

### OpenCode

Current OpenCode (research ref
`1f94d8a3c86b67f4f49a0e341de74e9188381b3a`) stores sessions in SQLite under
the platform XDG data directory, normally `opencode/opencode.db`; `OPENCODE_DB`
overrides the file. The canonical schema joins `session` to `project` and
provides:

- opaque session ID and slug;
- title;
- exact directory plus project worktree/relative path;
- created/updated epoch-millisecond timestamps;
- archive state and parent lineage.

Exact interactive resumption supports `opencode --session <id>`; the `run`
subcommand also accepts the same session ID.

## Local availability

Claude and Copilot stores are present on this machine. Hermes and OpenCode are
not installed, so their behavior must be proven with synthetic SQLite fixtures
and public-schema citations rather than local user data.

## Compatibility and risks

- Agent SQLite files are copied with their stable WAL/rollback journal to a
  private snapshot; scanners never SQLite-open, migrate, checkpoint, or mutate
  the source store. The private copy may recover its copied journal before
  `query_only=1` metadata queries.
- One malformed database/schema should fail that source explicitly, not silently
  produce an empty successful scan.
- Archived or internal lineage sessions should not clutter the default listing.
- Successful source snapshots must reconcile rows removed or archived upstream.
- Source-native exact IDs and verified paths take precedence over directory-name
  guesses.
- Transcript/message content remains lazy and out of scope for SQLite adapters.
- OpenCode channel-specific database discovery and legacy JSON stores are
  follow-up compatibility work; this slice supports the current stable DB plus
  explicit configuration/environment overrides.

# Historical session adapters

`tss scan` reads local agent stores into its own inventory. Source stores are
opened read-only and remain authoritative.

## Supported sources

| Source | Store | Inventory behavior | Exact resume |
| --- | --- | --- | --- |
| Claude Code | Direct-child `~/.claude/projects/*/*.jsonl` files or `CLAUDE_CONFIG_DIR` | Reads transcript metadata events for session ID, cwd, title, and timestamps. Subagent transcript directories are excluded. Raw conversation content remains lazy. | `claude --resume <id>` |
| GitHub Copilot CLI | `~/.copilot/session-state/*/workspace.yaml` or `COPILOT_HOME` | Reads workspace ID, cwd/Git root, repository, branch, name, and timestamps. | `copilot --resume=<id>` |
| Hermes Agent | `${HERMES_HOME}/state.db`, active profile DB, or `sources.hermes_database` | Queries a private DB+WAL/journal snapshot; lists unarchived logical CLI roots from metadata columns only. | `hermes --resume <id>` |
| OpenCode | `OPENCODE_DB`, XDG data DB, or `sources.opencode_database` | Queries a private DB+WAL/journal snapshot; joins non-archived root sessions to projects. | `opencode --session <id>` |

Herdr and tmux are separate live-runtime providers. They are not historical
transcript stores.

## Configuration

```yaml
sources:
  claude_projects: ~/.claude/projects
  copilot_session_state: ~/.copilot/session-state
  hermes_database: ~/.hermes/state.db
  opencode_database: ~/.local/share/opencode/opencode.db
```

Environment precedence:

- Claude: `CLAUDE_CONFIG_DIR`.
- Copilot: `COPILOT_HOME`.
- Hermes: `HERMES_HOME`; without it, an `active_profile` selects
  `~/.hermes/profiles/<name>/state.db`.
- OpenCode: `OPENCODE_DB`, then `XDG_DATA_HOME`, then
  `~/.local/share/opencode/opencode.db` on every platform, matching OpenCode's
  `xdg-basedir` dependency.
- Hermes on Windows defaults to `%LOCALAPPDATA%\hermes\state.db`.

Explicit config values override these defaults.

Generated resume commands preserve that exact store context:

```sh
CLAUDE_CONFIG_DIR='/custom/claude' claude --resume '<id>'
COPILOT_HOME='/custom/copilot' copilot --resume='<id>'
HERMES_HOME='/custom/hermes/profiles/coder' hermes --resume '<id>'
OPENCODE_DB='/custom/opencode.db' opencode --session '<id>'
```

When a root `HERMES_HOME` or the platform default contains a non-default
`active_profile`, `tss` resolves the same `<root>/profiles/<name>/state.db`
selected by the Hermes CLI bootstrap. An already profile-scoped
`HERMES_HOME=<root>/profiles/<name>` remains authoritative.
For an explicitly configured root `state.db`, `tss` also sets Hermes's
supervised-child bootstrap guard so a later sticky `active_profile` cannot
redirect the resume command to another database.

## Privacy boundary

- Scanners never write, migrate, checkpoint, repair, or prune source stores.
- Hermes/OpenCode DB+WAL or rollback-journal files are copied to a
  stability-checked private snapshot. SQLite may recover the private journal,
  then `query_only=1` prevents adapter writes. Source files are never
  SQLite-opened; coordination/recovery files exist only in temporary storage.
- Hermes message content, system prompts, gateway routing identity, and user/chat
  identifiers are not queried.
- OpenCode messages, parts, credentials, events, and prompts are not queried.
- Claude transcript content is not needed for inventory metadata; content
  remains available only to explicit lazy search/summarize workflows.
- Copilot `events.jsonl` and `session.db` are not read by the inventory scanner;
  `workspace.yaml` is the metadata authority.

## Default filtering

- Hermes: `source='cli'`, unarchived logical roots only; compression chains use
  their latest continuation metadata.
- OpenCode: non-archived root sessions with valid project relationships only.
- Claude/Copilot: one inventory row per native session/workspace ID.

Successful authoritative scans reconcile the local inventory: sessions deleted,
archived, or filtered out upstream disappear locally. Manual titles, pins, and
tags remain attached to sessions that still exist. Missing, failed, or malformed
source scans never prune prior inventory rows.

Hermes branch conversations are deferred; compression continuations are
projected into their stable root inventory entry. Archived sessions, internal
delegate/tool children, OpenCode legacy JSON, development-channel OpenCode
databases, Codex, and t3code are follow-up compatibility work.

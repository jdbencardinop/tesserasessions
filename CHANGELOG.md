# Changelog

All notable changes to `tesserasessions` will be documented in this file.

## Unreleased

- Bootstrap `tss`, a local-first CLI for inventorying coding-agent sessions.
- Add SQLite inventory storage, Claude/Copilot scanners, tmux/Herdr live scaffolding, local summaries, search, and manual session curation.
- Add the side-effect-free `tss status --json` batch runtime provider contract
  with explicit freshness, match evidence, raw observations, and separate
  runtime-presence and agent-state aggregates.
- Resolve linked worktrees to their common Git repository, reject contradictory
  branch hints, recognize the current Herdr schema, and exclude ordinary tmux
  shells from agent status.
- Reconcile persisted Herdr/tmux runtime snapshots after successful scans.
- Read exact Claude and Copilot workspace metadata, and add read-only current
  Hermes and OpenCode SQLite historical-session adapters.
- Exclude Claude subagent transcripts, project Hermes compression lineages,
  validate OpenCode relationships, reconcile authoritative history snapshots,
  and isolate SQLite WAL/rollback-journal reads and recovery in private
  snapshots.
- Preserve source-store environment in exact resume commands and continue
  Claude metadata scanning after oversized payload rows.
- Add a read-only OpenAI Codex CLI adapter for active JSONL rollout metadata,
  exact thread resume, interactive-source filtering, indexed names, and
  authoritative reconciliation.
- Add a mission-driven operator learning path with trusted resources, canonical
  terminology, a printable first-day runbook, and short interactive lessons.
- Add tpatch-tracked roadmap metadata and project documentation.

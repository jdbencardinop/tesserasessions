# Feature Request: Expand agent-source adapters beyond Claude and Copilot, starting with Hermes Agent and later t3code once the core inventory model is stable; keep parsers read-only and metadata-first.

**Slug**: `agent-source-expansion`
**Created**: 2026-07-07T04:31:46Z

## Description

Expand agent-source adapters beyond Claude and Copilot, starting with Hermes Agent and later t3code once the core inventory model is stable; keep parsers read-only and metadata-first.

Scope this feature to harden the existing Claude Code and GitHub Copilot CLI historical adapters and add current SQLite adapters for Hermes Agent and OpenCode. Preserve read-only, metadata-first behavior; use exact native session IDs and project paths; add synthetic fixtures because Hermes/OpenCode are not installed locally. Keep Codex and t3code adapters as later features after this four-source contract is stable.

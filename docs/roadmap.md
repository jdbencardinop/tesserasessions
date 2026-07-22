# Roadmap

This roadmap is tracked with Tessera Patch. Source of truth for feature state is `.tpatch/FEATURES.md`; this document explains the product intent and sequencing.

## Workflow

Before working on a tracked feature:

```sh
tpatch status <feature-slug>
tpatch next <feature-slug>
```

Then advance the feature through the lifecycle:

```text
requested -> analyzed -> defined -> implementing -> applied -> active
```

For this repo, `tpatch test <feature-slug>` runs:

```sh
go test ./...
```

## Feature DAG

```text
cli-inventory-core
|-- agent-source-expansion       (hard)
|-- docs-guide-cheatsheet        (soft)
|-- live-herdr-tmux-control      (hard)
|-- smart-session-summaries      (hard)
`-- session-ux-roadmap           (hard)
    `-- smart-session-summaries  (soft ordering hint)
```

## Tracked features

| Slug | Phase | Goal | Notes |
| --- | --- | --- | --- |
| `cli-inventory-core` | MVP foundation | Standalone Go CLI, SQLite inventory, Claude/Copilot scans, Herdr/tmux scaffolding, and basic commands. | Initial implementation exists; tpatch was adopted after the first slice, so record this feature before a real commit/landing flow. |
| `docs-guide-cheatsheet` | Docs slice | README, cheatsheet, scope, architecture, command examples, and tpatch workflow docs. | Depends softly on the core because docs can evolve while core changes. |
| `live-herdr-tmux-control` | Next live-control hardening | Improve live status matching, attach/open/send/run behavior, Herdr JSON support, and tmux pane targeting. | Herdr remains preferred; tmux remains fallback. |
| `smart-session-summaries` | Summary quality | Better local titles, goals, blockers, next actions, and confidence from recent transcript/pane output. | Remote LLM support must remain explicit opt-in. |
| `agent-source-expansion` | Adapter expansion | Add Hermes Agent first, then t3code once the core data model is stable. | Parsers should stay read-only and metadata-first. |
| `session-ux-roadmap` | Operator UX | Add project grouping, fuzzy/content search, tags, pin/done markers, stale thresholds, shell completions, and better filters. | Depends on stable inventory and summary metadata. |

## Milestones

### M0 - Working local inventory

Done enough for local use:

- `tss doctor`
- `tss scan`
- `tss list`
- `tss show`
- Claude and Copilot metadata adapters
- SQLite inventory
- Initial Herdr/tmux live adapter scaffolding
- Local/extractive summary command

### M1 - Documentation and tracked workflow

Current slice:

- README with scope and quick start.
- Cheatsheet for daily commands.
- Roadmap mapped to tpatch feature slugs.
- `.tpatch/` initialized with feature dependencies.
- `tpatch` test command set to `go test ./...`.

### M2 - Live control hardening

Target:

- Better matching between historical sessions and live Herdr/tmux panes.
- Safer `send` and `run` behavior. Initial Herdr `agent prompt`, tmux `send-keys`, and `run` command generation are implemented.
- `run <session> -- <command>` command. Initial Herdr/tmux implementation is added.
- `read <session>` command for recent live output is added.
- Clear status mapping: `needs_attention`, `working`, `idle`, `done`, `stale`, `unknown`.
- Herdr-first behavior when installed, tmux fallback when not.

### M3 - Summary quality

Target:

- Better first-prompt/latest-prompt extraction. Initial local candidate filtering and title/goal phrasing are implemented.
- Recent pane-output summarization for tmux and Herdr.
- Blocker and next-action fields.
- Confidence scoring based on source quality.
- Explicit opt-in remote summarizer command.

### M4 - More source adapters

Target:

- Hermes Agent state database/sessions.
- t3code after core schema stabilizes.
- Optional Codex/OpenCode adapter if local stores are present.
- Parser fixtures using redacted or synthetic data.

### M5 - Operator UX

Target:

- Tags and manual titles. Initial `title` and `tag` commands are implemented.
- Pin/done commands. Initial `pin` and `mark` commands are implemented.
- Project grouping and stale filters. Initial `--group-project`, `--sort`, `--tag`, and `--pinned` options are implemented.
- Fuzzy/content search. Initial `tss search`, `--content`, and optional `--fzf` are implemented.
- Shell completions.
- Better JSON output for scripting.

## Design principles

- Local first: no telemetry, no hosted database.
- Metadata first: avoid transcript parsing unless requested.
- Read-only source scans: do not mutate agent stores.
- Herdr first for live state; tmux fallback for portability.
- Explicit execution: support `--print` on potentially interactive commands.
- tpatch tracked: every roadmap slice should have a feature slug before implementation.

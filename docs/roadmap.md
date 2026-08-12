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
|-- docs-guide-cheatsheet        (soft)
|   `-- operator-learning-runbook (hard)
|-- live-herdr-tmux-control      (hard)
|   `-- runtime-status-contract  (hard)
|       `-- agent-source-expansion (soft ordering)
|           `-- codex-session-adapter (hard)
|-- smart-session-summaries      (hard)
`-- session-ux-roadmap           (hard)
    `-- smart-session-summaries  (soft ordering hint)
```

`runtime-status-contract` also depends directly on `cli-inventory-core`.
`agent-source-expansion` keeps its existing hard dependency on the core.
`operator-learning-runbook` also has a soft ordering edge from the latest
Codex adapter so lessons reflect the complete current source inventory.

## Delivered - runtime status contract

The side-effect-free, versioned status-provider contract for orchestrators such
as `tws` is implemented and verified. It does not replace the consumer's
topology, health, or rollup authority.

Cross-repository work is intentionally split:

1. `tws` implements `agent-work-status-dashboard` from its existing local
   signals without waiting for `tss`. Shipped in `tws` v1.2.12.
2. `tss` implements `runtime-status-contract`. Delivered.
3. `tws` later adds an optional `tss-status-enrichment` child feature with
   timeout/schema guards and baseline fallback.

See [Runtime status provider contract](runtime-status-contract.md).

## Delivered - agent source expansion

`agent-source-expansion` hardened the existing Claude/Copilot adapters and added
current Hermes/OpenCode SQLite stores while retaining the metadata-first,
read-only boundary.

## Delivered - Codex session adapter

`codex-session-adapter` adds active OpenAI Codex CLI rollout discovery through
bounded first-line metadata reads, exact thread identities, optional indexed
names, source filtering, deduplication, and source-qualified resume commands.
Archived/compressed rollouts and t3code remain separate work.

## Next - operator learning runbook

`operator-learning-runbook` is registered and unblocked. It adapts the
[teach skill](https://github.com/mattpocock/skills/tree/main/skills/productivity/teach)
into a `tss` operator curriculum: a concrete mission, trusted annotated
resources, canonical terminology, printable references, short practice
lessons, and a policy for local learning records. Exercises must use real `tss`
commands with privacy-safe synthetic data.

## Tracked features

| Slug | Phase | Goal | Notes |
| --- | --- | --- | --- |
| `cli-inventory-core` | MVP foundation | Standalone Go CLI, SQLite inventory, Claude/Copilot scans, Herdr/tmux scaffolding, and basic commands. | Initial implementation exists; tpatch was adopted after the first slice, so record this feature before a real commit/landing flow. |
| `docs-guide-cheatsheet` | Docs slice | README, cheatsheet, scope, architecture, command examples, and tpatch workflow docs. | Depends softly on the core because docs can evolve while core changes. |
| `live-herdr-tmux-control` | Live-control foundation | Improve live status matching, attach/open/send/run behavior, Herdr JSON support, and tmux pane targeting. | Applied; Herdr remains preferred and tmux remains fallback. |
| `runtime-status-contract` | Status provider | Publish a side-effect-free batch JSON provider with freshness, match evidence, raw runtimes, and separate presence/agent-state aggregates. | Applied and verified; does not own the `tws` rollup and never guesses cwd-less sessions. |
| `smart-session-summaries` | Summary quality | Better local titles, goals, blockers, next actions, and confidence from recent transcript/pane output. | Remote LLM support must remain explicit opt-in. |
| `agent-source-expansion` | Historical adapter expansion | Harden Claude/Copilot and add current Hermes/OpenCode historical adapters. | Applied; read-only and metadata-first. |
| `codex-session-adapter` | Codex historical adapter | Add active Codex JSONL rollout metadata, names, exact resume, filtering, and reconciliation. | Applied; archived and compressed rollouts remain deferred. |
| `operator-learning-runbook` | Operator onboarding | Build a mission-driven runbook and progressive learning workspace for discovering, curating, resuming, monitoring, and troubleshooting sessions. | Requested and unblocked; adapt the teach skill rather than copying its workspace layout blindly. |
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

### M3 - Runtime status provider contract

Delivered:

- Versioned batch request/response JSON on stdin/stdout.
- Side-effect-free live probes with no implicit scan.
- Opaque query IDs plus verified path and optional Git hints.
- Separate `runtime_presence` and `agent_state`.
- Provider availability, observation timestamps, expiry, and per-query errors.
- Raw runtime observations plus deterministic aggregation.
- Explicitly unmatched cwd-less sessions.
- Runtime-row reconciliation during explicit scans.

The feature is unblocked: both hard dependencies, `cli-inventory-core` and
`live-herdr-tmux-control`, are applied.

### M4 - Summary quality

Target:

- Better first-prompt/latest-prompt extraction. Initial local candidate filtering and title/goal phrasing are implemented.
- Recent pane-output summarization for tmux and Herdr.
- Blocker and next-action fields.
- Confidence scoring based on source quality.
- Explicit opt-in remote summarizer command.

### M5 - More source adapters

Delivered:

- Exact Claude/Copilot metadata and resume identities.
- Hermes Agent state database/sessions.
- OpenCode SQLite sessions/projects.
- Active Codex JSONL rollout metadata and exact thread resume.
- Parser fixtures using redacted or synthetic data.

Remaining:

- t3code application/server sessions as a separate feature after Codex or when
  the desktop/server inventory contract is defined.

### M6 - Operator UX

Target:

- Tags and manual titles. Initial `title` and `tag` commands are implemented.
- Pin/done commands. Initial `pin` and `mark` commands are implemented.
- Project grouping and stale filters. Initial `--group-project`, `--sort`, `--tag`, and `--pinned` options are implemented.
- Fuzzy/content search. Initial `tss search`, `--content`, and optional `--fzf` are implemented.
- Shell completions.
- Better JSON output for scripting.

### M7 - Operator learning

Next:

- Define an observable operator mission and success criteria.
- Curate annotated primary resources for `tss` and each supported integration.
- Establish a canonical glossary for inventory, runtime, and agent-state terms.
- Produce a printable first-day runbook and troubleshooting references.
- Add short, reusable practice lessons with privacy-safe synthetic fixtures.
- Keep individual learning evidence local; commit only reusable formats and
  generic course material.

## Design principles

- Local first: no telemetry, no hosted database.
- Metadata first: avoid transcript parsing unless requested.
- Read-only source scans: do not mutate agent stores.
- Status-provider queries are side-effect-free and never imply a scan.
- External consumers use versioned JSON, never SQLite or Go imports.
- Consumers retain topology, local-health, and final-rollup authority.
- Herdr first for live state; tmux fallback for portability.
- Explicit execution: support `--print` on potentially interactive commands.
- tpatch tracked: every roadmap slice should have a feature slug before implementation.

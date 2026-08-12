# Mission: Operate tss with confidence

## Why

Use `tss` to recover the right coding-agent work quickly, judge what evidence
is current, and return to it safely without exposing transcript content or
confusing historical inventory with live runtime state.

## Success looks like

- Build an inventory and explain whether each source was found, skipped,
  incomplete, or failed.
- Find the intended session from metadata and inspect it before taking action.
- Preview and choose between attaching, resuming, or opening a new workspace.
- Explain `runtime_presence` separately from `agent_state` and identify when
  local evidence needs attention.
- Leave useful titles, tags, pins, and summaries for the next handoff, and
  understand when a later scan can recompute status.

## Constraints

- Prefer metadata and local/extractive behavior over transcript access.
- Preview interactive commands with `--print`.
- Use official contracts and current command help as the source of truth.
- Practice with synthetic IDs and paths; never commit local inventory data.
- Keep each lesson focused enough to finish in about ten minutes.

## Out of scope

- Learning how to operate each underlying coding agent.
- Editing agent source stores or the `tss` SQLite database.
- Hosted progress tracking, telemetry, or synchronized learner profiles.
- Treating a completed lesson as proof of understanding without evidence.

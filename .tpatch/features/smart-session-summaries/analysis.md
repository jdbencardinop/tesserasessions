# Analysis: smart-session-summaries

## Summary

Improve local, privacy-preserving session summaries so `tss summarize` produces better titles and goal summaries from existing transcript or pane-output data without calling a remote LLM.

## Compatibility

- Compatible with current SQLite schema because summaries already store title, goal summary, provider, confidence, and timestamp.
- Content extraction can remain lazy and read-only.
- Existing CLI contract stays stable: `tss summarize <session>` updates title and goal summary.

## Affected areas

- `internal/summarize/summarize.go`
- `internal/adapters/text.go`
- `internal/adapters/claude.go`
- `internal/store/store.go` if UX fields need to preserve manual titles later
- tests under `internal/summarize` and `internal/adapters`

## Risks

- Transcript structures vary between agents.
- Local extractive summaries can be noisy if tool output dominates the transcript.
- Summary updates should not overwrite manually curated titles once manual-title UX lands.

# Analysis: session-ux-roadmap

## Summary

Improve operator UX for managing many sessions by adding better list controls, manual curation fields, and fuzzy/content search. Fuzzy search belongs in this feature because it is an interaction pattern layered on top of inventory and summaries.

## Compatibility

- Requires backward-compatible SQLite migrations for optional UX fields.
- Must not break existing `scan`, `list`, or `show` output.
- fzf must remain optional; non-interactive search should work without it.

## Affected areas

- `internal/store/store.go`
- `internal/core/model.go`
- `internal/cli/root.go`
- `internal/adapters/text.go`
- docs for command usage

## Risks

- Content search can be slow if it scans many large transcript files; it should be opt-in.
- fzf may not be installed; command should return a helpful error or fall back cleanly.
- Manual titles should not be overwritten by later scanner runs.

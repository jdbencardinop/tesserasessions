# Exploration: smart-session-summaries

## Relevant files

- `internal/summarize/summarize.go`: local summary logic.
- `internal/adapters/text.go`: generic text extraction used by summaries and future search.
- `internal/adapters/claude.go`: JSONL reader helper.
- `internal/core/model.go`: string truncation and status constants.

## Minimal changeset

- Add candidate scoring/filtering in the summarizer.
- Expand text candidate reading to simple text and JSON-like files for content search reuse.
- Preserve existing CLI command names and DB shape where possible.

## Validation

- `go test ./...`
- `tss summarize <known-session>` manually on indexed data.

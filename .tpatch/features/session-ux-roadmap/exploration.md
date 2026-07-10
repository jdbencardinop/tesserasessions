# Exploration: session-ux-roadmap

## Relevant files

- `internal/store/store.go`: filtering, schema, updates.
- `internal/core/model.go`: session fields.
- `internal/cli/root.go`: commands and output.
- `internal/adapters/text.go`: lazy content snippets.
- `README.md` and `docs/cheatsheet.md`: command docs.

## Minimal changeset

- Add columns: `title_source`, `pinned`, `tags`.
- Preserve manual titles during scanner upserts.
- Add CLI commands for manual curation and search.
- Keep fzf optional via `exec.LookPath("fzf")`.

## Validation

- `go test ./...`
- `tss list --query <term>` includes summaries.
- `tss search <term>` returns metadata matches.
- `tss search <term> --content` can scan lazy content.
- `tss search --fzf --limit 20` errors clearly if fzf is unavailable.

# Spec: session-ux-roadmap

## Problem

As the inventory grows, simple `list` output is not enough. Users need sorting, grouping, manual curation, and search/fuzzy picking across title, summary, metadata, and optionally content.

## Acceptance criteria

1. `tss list` can sort and group sessions by project.
2. Query filters search title, summary, project path, native id, and agent metadata.
3. Sessions can be manually titled, marked with status, pinned/unpinned, and tagged.
4. `tss search` provides non-interactive search and optional fzf selection.
5. Content search is opt-in and lazy.
6. Manual titles survive future scans.
7. `go test ./...` passes.

## Out of scope

- Full-text indexing/FTS database in this slice.
- Rich TUI.
- Hosted sync.

## Plan

1. Add backward-compatible UX columns to the session table.
2. Add store update methods for title, status, pin, and tags.
3. Extend list filters/sorting/grouping.
4. Add `search`, `title`, `mark`, `pin`, and `tag` commands.
5. Add optional `--fzf` mode to search.
6. Update docs.

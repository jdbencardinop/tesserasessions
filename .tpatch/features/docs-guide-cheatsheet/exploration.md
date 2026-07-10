# Exploration: docs-guide-cheatsheet

## Relevant files

- `cmd/tss/main.go`: CLI entry point.
- `internal/cli/root.go`: command definitions, flags, and help text.
- `internal/config/config.go`: config/data defaults and environment overrides.
- `internal/adapters/*.go`: supported source and live adapters.
- `internal/store/store.go`: SQLite inventory schema.
- `internal/summarize/summarize.go`: local/extractive summary behavior.
- `.tpatch/FEATURES.md`: generated tracked feature list.
- `.tpatch/config.yaml`: tpatch provider/test/dependency config.

## Current command surface

- `tss doctor`
- `tss scan`
- `tss list` / `tss ls`
- `tss show <session>`
- `tss summarize <session>`
- `tss attach <session>`
- `tss open <session>`
- `tss send <session> <text>`

## Minimal changeset

- Add `README.md`.
- Add `docs/cheatsheet.md`.
- Add `docs/roadmap.md`.
- Keep docs consistent with `go run ./cmd/tss --help`.

## Validation

- `go test ./...`
- `tpatch feature deps --validate-all`
- `tpatch status --dag`


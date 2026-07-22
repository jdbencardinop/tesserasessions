# Exploration: repo-scaffold-consistency

## Relevant sibling files

- `../tesserapatch/tpatch/Makefile`
- `../tesserapatch/tpatch/internal/buildinfo/buildinfo.go`
- `../tesserapatch/tpatch/internal/buildinfo/buildinfo_test.go`
- `../tesserapatch/tpatch/internal/cli/cobra.go`
- `../tesserapatch/tpatch/.gitignore`
- `../tesserapatch/tpatch/LICENSE`
- `../tesseraspaces/Makefile`
- `../tesseraspaces/cmd/tws/main.go`
- `../tesseraspaces/internal/cli/root.go`
- `../tesseraspaces/README.md`
- `../tesseraspaces/LICENSE`

## Relevant local files

- `cmd/tss/main.go`
- `internal/cli/root.go`
- `README.md`
- `.gitignore`
- `go.mod`

## Proposed local files

- `Makefile`
- `LICENSE`
- `CHANGELOG.md`
- `internal/buildinfo/buildinfo.go`
- `internal/buildinfo/buildinfo_test.go`

## Notes

Prefer the `tesserapatch` `internal/buildinfo` approach over `tesseraspaces`'s `main.version` approach because it keeps version resolution reusable and testable outside `main`.


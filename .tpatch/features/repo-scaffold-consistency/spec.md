# Spec: repo-scaffold-consistency

## Problem

`tesserasessions` should feel like part of the same Tessera CLI family as `tesserapatch` and `tesseraworkspaces`: same Makefile ergonomics, version injection strategy, standard project metadata, and README sections.

## Acceptance criteria

1. Add a `Makefile` with `build`, `test`, `fmt`, `lint`, `install`, `clean`, and `all`.
2. Build output goes to `bin/tss`.
3. Version is resolved from `git describe --tags --always --dirty` in Makefile builds.
4. CLI exposes version through Cobra using an `internal/buildinfo` package.
5. BuildInfo fallback works for module-installed versions and `dev` builds.
6. Add tests for buildinfo version resolution.
7. Add `LICENSE` with the same license family as sibling projects.
8. Add `CHANGELOG.md` scaffold.
9. Expand `.gitignore` to match sibling Go conventions while preserving `.gitignored/`.
10. Update README install-from-source/build instructions to reference `make`.
11. `go test ./...` passes.

## Out of scope

- Release automation.
- Homebrew formula.
- GitHub Actions.
- Changing package/module name.

## Plan

1. Add `internal/buildinfo` package and tests, modeled after `tesserapatch`.
2. Wire Cobra root `Version` to `buildinfo.String()`.
3. Add Makefile with version ldflags.
4. Add license/changelog scaffolds.
5. Expand README and `.gitignore`.
6. Validate with `make test`, `make build`, and `bin/tss --version`.


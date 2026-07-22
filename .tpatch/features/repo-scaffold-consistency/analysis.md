# Analysis: repo-scaffold-consistency

## Summary

Align `tesserasessions` with the sibling Tessera Go CLI repositories so build, install, version output, docs, and project metadata feel consistent. The requested comparison targets were `../tesserapatch` and `../tesseraworkspaces`; in this workspace, `../tesseraworkspaces` does not exist, and the matching sibling is `../tesseraspaces`, whose module and README are named `tesseraworkspaces`.

## Findings from `../tesserapatch/tpatch`

- Uses a `Makefile` with targets: `build`, `test`, `fmt`, `lint`, `install`, `clean`, `all`.
- Builds to `bin/tpatch`.
- Injects version with:
  - `VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)`
  - `-ldflags "-X github.com/tesseracode/tesserapatch/internal/buildinfo.Version=$(VERSION)"`
- Centralizes version resolution in `internal/buildinfo`.
- `internal/buildinfo.String()` prefers ldflags, then `runtime/debug.ReadBuildInfo()`, then `dev`.
- Cobra root uses `Version: buildinfo.String()` and a custom version template.
- `.gitignore` is based on github/gitignore Go rules and avoids shadowing `cmd/tpatch/` with an anchored `/tpatch` binary ignore.
- Includes `LICENSE`, `CHANGELOG.md`, `README.md`, `AGENTS.md`, `CLAUDE.md`, `SPEC.md`, and extensive `docs/`.

## Findings from `../tesseraspaces` (`tesseraworkspaces`)

- Uses a similar `Makefile` target set: `build`, `test`, `fmt`, `lint`, `install`, `clean`, `all`.
- Builds to `bin/tws`.
- Injects version with:
  - `VERSION=$(shell git describe --tags --always --dirty 2>/dev/null || echo dev)`
  - `-ldflags "-X main.version=$(VERSION)"`
- `cmd/tws/main.go` has a `version` var, resolves module build info fallback, calls `cli.SetVersion(...)`, and exits with `cli.Execute()`.
- `internal/cli/root.go` exposes Cobra `Version: version`.
- README has a consistent shape: quick start, how it works, features, all commands, requirements, configuration, documentation, install from source, license.
- Includes `LICENSE`, `CHANGELOG.md`, `Makefile`, `docs/`, and agent skill files.

## Current gaps in `tesserasessions`

- No `Makefile`.
- No `LICENSE`.
- No `CHANGELOG.md`.
- No CLI `--version` wiring.
- No centralized `internal/buildinfo` package.
- README exists but can be aligned more closely with sibling structure and install-from-source commands.
- `.gitignore` exists but is minimal compared with sibling Go gitignores.

## Compatibility

This is a scaffolding/documentation feature. It should not change inventory behavior except for adding `tss --version` / `tss version` support through Cobra's built-in version flag.


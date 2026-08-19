# WSL handoff

Read this first when continuing `tesserasessions` from WSL. This is a temporary
cross-machine handoff; delete it after the WSL session has absorbed the context.

## Current state

- Repository: `https://github.com/jdbencardinop/tesserasessions` (private).
- Branch: `main`, synchronized with `origin/main`.
- Baseline before this handoff: `5340f10`.
- Product: local-first `tss` CLI for coding-agent historical inventory, search,
  curation, native resume, and Herdr/tmux runtime observation/control.
- Delivered historical sources: Claude Code, GitHub Copilot CLI, Hermes Agent,
  OpenCode, and OpenAI Codex CLI.
- Delivered external contract: side-effect-free, versioned
  `tss status --json`.
- Delivered operator course: `docs/learn/`.
- Next product backlog item is a separate t3code adapter, but do not start it
  unless the user explicitly moves on from WSL validation and learning.

The current ownership decision is:

- Tesserabot owns its multi-user chats, messages, LLM/tool runs, persistence,
  authentication, and search.
- `tss` owns one OS user's local coding-harness inventory and runtime state.
- `tws` owns worktree/workspace topology and sessions it launches.
- Do not implement a Tesserabot integration in this repository. Its remaining
  topology investigation is tracked separately in
  `jdbencardinop/tesserabot#1`.

## Important WSL boundary

The Git repository synchronizes code and documentation. Agent stores and the
`tss` inventory do not synchronize through Git.

Inside WSL:

- the default inventory is
  `~/.local/share/tesserasessions/sessions.db` inside that WSL distro;
- source discovery defaults to the WSL user's Linux home;
- Linux tmux/Herdr processes are distinct from Windows or macOS processes;
- a Windows-native agent store is not automatically a WSL-native store;
- paths under `/mnt/c` are visible files, but do not assume Windows-native
  process state, permissions, locking, or SQLite sidecar behavior is equivalent.

Establish a WSL-native baseline first. Do not point adapters at Windows stores
under `/mnt/c` until the default tests and smoke checks pass, and treat any such
experiment as a separate compatibility observation.

For meaningful source smoke testing, run the agent and `tss` in the same WSL
environment. A missing source or live backend is an honest `skipped` or
unavailable result, not a reason to fabricate fixtures outside tests.

## Restore the repository

Authenticate GitHub CLI if needed:

```sh
gh auth login
gh repo clone jdbencardinop/tesserasessions
cd tesserasessions
git pull --ff-only
git status --short --branch
```

If the repository already exists:

```sh
cd /path/to/tesserasessions
git pull --ff-only
git status --short --branch
```

The tree should be clean and `main` should track `origin/main`.

## Toolchain

Required by the repository:

- Go 1.26.4 or newer;
- Git;
- Make.

Useful for WSL validation:

- `build-essential` for the Go race detector;
- `tmux`;
- `sqlite3`;
- `python3`;
- GitHub CLI (`gh`);
- optional `wslu` for `wslview`.

On Ubuntu WSL, install the distro packages that are available:

```sh
sudo apt update
sudo apt install -y build-essential git make tmux sqlite3 python3
```

Install the required Go version from an official or already trusted toolchain
manager if `go version` is older than the module requirement:

```sh
go version
```

Do not lower the `go` version in `go.mod` to accommodate the machine.

## Required WSL validation

Run from the repository root:

```sh
make test
make lint
make build
go test -race ./...
git diff --check
```

Then run the CLI smoke checks:

```sh
./bin/tss --version
./bin/tss --help
./bin/tss doctor
./bin/tss scan --json
./bin/tss list --json
```

`scan` writes only the WSL-local `tss` inventory and reads configured source
stores. It does not mutate source stores. On a fresh WSL distro, zero sessions
and skipped sources are expected.

Exercise the live status contract without mutating the inventory:

```sh
python3 -c 'import json, os; print(json.dumps({
  "schema_version": 1,
  "queries": [{"query_id": "wsl-repo", "path": os.getcwd()}]
}))' |
  ./bin/tss status --json |
  python3 -m json.tool
```

This must return valid schema-v1 JSON even when Herdr or tmux is unavailable.
Provider unavailability should remain explicit rather than becoming a false
`active` or `absent` result.

Also capture:

```sh
uname -a
cat /etc/os-release
pwd
df -T .
go version
command -v tmux || true
command -v herdr || true
```

Report failures with the command and redacted output. Do not attach databases,
agent transcripts, credentials, home-directory listings, or raw source stores.

## WSL-specific observations wanted

Please report:

1. Whether all repository, race, and build gates pass on the WSL distro.
2. Whether the checkout is on the WSL Linux filesystem or `/mnt/c`; prefer the
   Linux filesystem for builds and SQLite behavior.
3. Which default source paths `tss doctor` finds or marks missing.
4. Whether WSL-native tmux or Herdr is installed and observable.
5. Whether a WSL-native agent session can be scanned and resumed with the exact
   source-qualified command.
6. Any path, permission, symlink, file-locking, case-sensitivity, browser-open,
   or shell-quoting difference.
7. Whether generated commands use Linux paths and quoting correctly.

Do not treat inability to observe a Windows-native process from WSL as a `tss`
bug. Record it as a topology boundary unless a WSL-native process/store also
fails.

## Open the learning course

Read the mission first:

```sh
less docs/learn/MISSION.md
```

Then open the lessons in this order:

1. `docs/learn/lessons/0001-build-an-inventory.html`
2. `docs/learn/lessons/0002-find-the-right-session.html`
3. `docs/learn/lessons/0003-return-safely.html`
4. `docs/learn/lessons/0004-diagnose-live-state.html`

If `wslview` is installed:

```sh
wslview "$PWD/docs/learn/lessons/0001-build-an-inventory.html"
```

Without `wslview`, ask Windows to open the local file:

```sh
powershell.exe -NoProfile -Command \
  "Start-Process '$(wslpath -w "$PWD/docs/learn/lessons/0001-build-an-inventory.html")'"
```

The most reliable alternative is a temporary local server:

```sh
python3 -m http.server 8000 --directory docs/learn
```

Then open this from the Windows browser:

```text
http://localhost:8000/lessons/0001-build-an-inventory.html
```

Or launch it from another WSL shell:

```sh
powershell.exe -NoProfile -Command \
  "Start-Process 'http://localhost:8000/lessons/0001-build-an-inventory.html'"
```

Stop the temporary server with `Ctrl-C`. It is only a static file server; the
course itself needs no hosted service.

After the four lessons, open or print:

```sh
wslview "$PWD/docs/learn/reference/first-day-runbook.html"
```

Use `docs/learn/GLOSSARY.md` and `docs/learn/RESOURCES.md` as references rather
than reading assignments.

## Development workflow

This repository uses Tessera Patch Path B:

1. Run `tpatch status <slug>`.
2. Run `tpatch next <slug> --format harness-json`.
3. Author analysis/spec/exploration/recipe directly.
4. Advance phases with `--manual`; never call the configured tpatch AI
   provider.
5. Run independent review loops and project gates.
6. Record before committing, then land and push.

Known limitation: `tpatch verify` recipe replay can pass while post-apply
closure replay fails after landing through historical parent features. Preserve
the truthful failure overlay; it is tracked in
`tesseracode/tesserapatch#8`.

Do not rewrite Git history, publish the private repository, commit local
inventories/source stores, or start unrelated Tesserabot integration work.

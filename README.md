# tesserasessions

`tss` is a local-first CLI for inventorying coding-agent sessions across tools like GitHub Copilot CLI, Claude Code, Herdr, and tmux.

It answers: "What agent sessions exist, what project are they for, when did they last move, and how do I get back to them?"

## Scope

The MVP is an inventory and control layer, not a replacement for agents or terminal multiplexers.

In scope:

- Scan local agent session stores in read-only mode.
- Normalize sessions into a local SQLite inventory.
- List, filter, and inspect sessions.
- Prefer Herdr for live control when available; fall back to tmux.
- Generate local/extractive summaries by default.
- Track roadmap features with Tessera Patch (`tpatch`).

Out of scope for now:

- Full GUI.
- Hosted sync or telemetry.
- Automatic remote LLM summarization.
- Deep t3code parsing before the core adapters are stable.
- Replacing Herdr, tmux, Agent Sessions, or AgentTrace.

## Requirements

- Go 1.26+
- git
- tmux (optional, for tmux live-session fallback)
- Herdr (optional, preferred live-session backend)

## Install from source

```sh
git clone https://github.com/jdbencardinop/tesserasessions.git
cd tesserasessions
make install
```

For local development:

```sh
make build        # writes bin/tss with git-derived version metadata
make test         # runs go test ./...
make lint         # gofmt check + go vet
go run ./cmd/tss --help
bin/tss --version
```

## Quick start

```sh
# Check config, source stores, optional backends, and inventory count.
go run ./cmd/tss doctor

# Scan all known sources.
go run ./cmd/tss scan

# List recent sessions.
go run ./cmd/tss list

# Inspect one session.
go run ./cmd/tss show <session-id>

# Print the command that would attach/resume.
go run ./cmd/tss attach <session-id> --print

# Create/update a local summary.
go run ./cmd/tss summarize <session-id>
```

Once installed as `tss`, drop the `go run ./cmd/tss` prefix:

```sh
tss scan
tss list --source claude --limit 20
tss show claude-abc123
```

## Commands

| Command | Purpose |
| --- | --- |
| `tss doctor` | Show config paths, source health, backend availability, and inventory count. |
| `tss scan` | Refresh the local inventory from configured sources. |
| `tss list` / `tss ls` | List indexed sessions with filters. |
| `tss show <session>` | Show metadata, paths, summary, and resume/attach commands. |
| `tss summarize <session>` | Generate a local/extractive title and goal summary. |
| `tss search [query]` | Search by metadata/summary and optionally content; supports fzf selection. |
| `tss title <session> <title>` | Set a manual title that future scans preserve. |
| `tss mark <session> <status>` | Set a normalized session status. |
| `tss pin <session>` | Pin a session so it sorts first. |
| `tss tag <session> <tag[,tag...]>` | Replace session tags. |
| `tss attach <session>` | Attach to a live runtime or run a native resume command. |
| `tss open <session>` | Open a new Herdr workspace or tmux session in the same directory. |
| `tss send <session> <text>` | Send text to a live Herdr/tmux runtime. |

Useful flags:

```sh
tss scan --source claude
tss list --source copilot --limit 10
tss list --query tesseraspaces
tss list --status needs_attention
tss list --group-project --sort project
tss list --pinned
tss search auth --content
tss search --fzf --show
tss title <session> "Refactor auth middleware"
tss mark <session> done
tss pin <session>
tss tag <session> mvp,review
tss attach <session> --backend tmux --print
tss open <session> --backend herdr --print
tss --db /tmp/tss.db scan
```

## How it works

`tss scan` runs source adapters and stores normalized records in SQLite.

Initial adapters:

- `claude`: scans `~/.claude/projects` or `CLAUDE_CONFIG_DIR/projects`.
- `copilot`: scans `~/.copilot/session-state` or `COPILOT_HOME/session-state`.
- `herdr`: uses `herdr agent list --json` when Herdr is installed and running.
- `tmux`: uses `tmux list-panes -a` when a tmux server is running.

The inventory database lives at:

```text
~/.local/share/tesserasessions/sessions.db
```

The default config path is:

```text
~/.config/tesserasessions/config.yaml
```

Overrides:

```sh
TSS_CONFIG=/path/to/config.yaml tss scan
TSS_DATA_DIR=/path/to/data tss scan
tss --db /path/to/sessions.db list
```

## Privacy and safety

- Scanners are read-only against agent session stores.
- Transcript parsing is lazy and only used where needed, such as `summarize`.
- Content search is opt-in with `tss search --content`.
- fzf selection is optional with `tss search --fzf`; non-interactive search works without fzf.
- Summaries are local/extractive by default.
- Remote LLM summarization is intentionally not automatic and should remain explicit opt-in.
- `attach`, `open`, and `send` support `--print` so commands can be reviewed before execution.

## Tessera Patch workflow

This repo uses `tpatch` to track feature work.

Common commands:

```sh
tpatch status --dag
tpatch next <feature-slug>
tpatch analyze <feature-slug>
tpatch define <feature-slug>
tpatch explore <feature-slug>
tpatch implement <feature-slug>
tpatch apply <feature-slug>
tpatch test <feature-slug>
tpatch record <feature-slug>
```

Configured test command:

```sh
go test ./...
```

Tracked feature slugs are listed in `.tpatch/FEATURES.md` and summarized in [docs/roadmap.md](docs/roadmap.md).

## More docs

- [Cheatsheet](docs/cheatsheet.md)
- [Roadmap](docs/roadmap.md)

## License

[MIT](LICENSE)

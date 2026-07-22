# tss cheatsheet

## Daily loop

```sh
tss doctor
tss scan
tss list --limit 20
tss show <session-id>
tss attach <session-id> --print
```

## Find sessions

```sh
tss list
tss list --source claude
tss list --source copilot
tss list --query tesseraspaces
tss list --status needs_attention
tss list --tag review
tss list --pinned
tss list --sort project --group-project
tss list --sort title
tss list --json
```

## Search and fuzzy pick

```sh
tss search auth
tss search auth --source claude
tss search auth --content
tss search --fzf
tss search auth --fzf --show
```

Search checks title, summary, project path, native id, agent, and tags. `--content` lazily scans supported raw files/directories. `--fzf` requires `fzf` on `PATH` and prints the selected session id unless `--show` is used.

## Inspect one session

```sh
tss show <session-id>
tss show <session-id-prefix>
tss show <native-session-id>
```

Session IDs are stable hashes like:

```text
claude-5d96f6be2a1e
copilot-25e90acc1fc7
tmux-...
herdr-...
```

## Refresh inventory

```sh
tss scan
tss scan --source claude
tss scan --source copilot
tss scan --source herdr
tss scan --source tmux
tss scan --json
```

## Summarize locally

```sh
tss summarize <session-id>
```

Current behavior:

- Reads recent transcript candidates for supported historical sources.
- Uses local/extractive heuristics.
- Updates title, goal summary, and sometimes status.
- Preserves manual titles set with `tss title`.
- Does not call a remote LLM by default.

## Curate sessions

```sh
tss title <session-id> "Fix auth middleware"
tss mark <session-id> needs_attention
tss mark <session-id> done
tss pin <session-id>
tss pin <session-id> --unpin
tss tag <session-id> mvp,review
```

Valid statuses:

```text
needs_attention working idle done stale unknown
```

## Attach, open, and send

Prefer review mode first:

```sh
tss attach <session-id> --print
tss open <session-id> --print
tss send <session-id> "continue from the latest TODO" --print
```

Run the command directly:

```sh
tss attach <session-id>
tss open <session-id>
tss send <session-id> "please summarize your current blocker"
```

Backend selection:

```sh
tss attach <session-id> --backend herdr
tss attach <session-id> --backend tmux
tss open <session-id> --backend tmux
tss send <session-id> --backend herdr "status?"
```

Default backend order:

1. Herdr, when installed and a matching live runtime exists.
2. tmux fallback.
3. Native resume command, when the historical adapter knows one.

## Config and data

Default config:

```text
~/.config/tesserasessions/config.yaml
```

Default database:

```text
~/.local/share/tesserasessions/sessions.db
```

Override paths:

```sh
TSS_CONFIG=/tmp/tss.yaml tss scan
TSS_DATA_DIR=/tmp/tss-data tss scan
tss --db /tmp/tss.db list
```

## Source adapters

| Source | What it scans |
| --- | --- |
| `claude` | `~/.claude/projects` or `CLAUDE_CONFIG_DIR/projects` |
| `copilot` | `~/.copilot/session-state` or `COPILOT_HOME/session-state` |
| `herdr` | `herdr agent list --json` |
| `tmux` | `tmux list-panes -a` |

## tpatch feature workflow

Check roadmap state:

```sh
tpatch status --dag
tpatch status <feature-slug>
tpatch next <feature-slug>
```

Normal lifecycle:

```sh
tpatch analyze <feature-slug>
tpatch define <feature-slug>
tpatch explore <feature-slug>
tpatch implement <feature-slug>
tpatch apply <feature-slug>
tpatch test <feature-slug>
tpatch record <feature-slug>
```

Useful dependency commands:

```sh
tpatch feature deps <feature-slug>
tpatch feature deps --validate-all
```

Feature files live under:

```text
.tpatch/features/<feature-slug>/
```

## Development checks

```sh
make test
make build
make lint
go test ./...
go build -o bin/tss ./cmd/tss
bin/tss --version
```

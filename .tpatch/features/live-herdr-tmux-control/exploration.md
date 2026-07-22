# Exploration: live-herdr-tmux-control

## Relevant files

- `internal/adapters/herdr.go`: Herdr runtime scanner and JSON parsing helpers.
- `internal/adapters/tmux.go`: tmux runtime scanner shape.
- `internal/cli/root.go`: attach/open/send command generation; add read/run commands here.
- `internal/store/store.go`: runtime lookup already exists.
- `README.md` and `docs/cheatsheet.md`: command docs.

## Herdr command targets

- `herdr agent list --json`: primary scanner source.
- `herdr agent attach <target>`: attach to live agent.
- `herdr agent prompt <target> <text>`: send prompt/text.
- `herdr agent read <target> --source recent --lines N`: read recent output.
- `herdr pane split <pane_id> --direction right --cwd PATH --no-focus`: create a pane.
- `herdr pane run <pane_id> <command>`: run command in a pane.

## tmux command targets

- `tmux list-panes -a`: scanner source.
- `tmux attach -t <session>`: attach.
- `tmux send-keys -t <pane> <text> Enter`: send.
- `tmux capture-pane -p -t <pane> -S -N`: read.
- `tmux split-window -h -t <pane> -c <path> <command>`: run command in a split.

## Validation

- `go test ./...`
- `make lint`
- disposable tmux scan/show/read/run `--print` smoke checks.
- Herdr command generation tests without requiring Herdr installation.

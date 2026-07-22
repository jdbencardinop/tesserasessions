# Spec: live-herdr-tmux-control

## Problem

`tss` has historical inventory and basic live scaffolding, but live sessions need practical controls: robust Herdr discovery, prompt/send, read, and run-command behavior, with tmux as fallback.

## Acceptance criteria

1. Herdr scanner handles nested JSON payloads such as `{ "result": { "agents": [...] } }`.
2. Herdr runtime records prefer usable agent targets and keep pane IDs for pane-level operations.
3. `tss send` uses `herdr agent prompt` for Herdr runtimes and `tmux send-keys` for tmux runtimes.
4. Add `tss read <session>` with Herdr and tmux implementations plus `--print`.
5. Add `tss run <session> -- <command>` that opens/runs in a Herdr split or tmux split when a live runtime exists.
6. Existing `attach`, `open`, and tmux scan behavior remain compatible.
7. `go test ./...`, `make lint`, and command-generation smoke checks pass.

## Out of scope

- Raw Herdr socket client.
- Full `session.snapshot` cache.
- Installing Herdr integrations automatically.
- Perfect handling of every possible Herdr JSON schema.

## Plan

1. Improve Herdr JSON object extraction and field lookup helpers.
2. Store Herdr pane ID in runtime `Surface` and agent target in `NativeID`.
3. Update send command generation.
4. Add `read` and `run` commands.
5. Add unit tests for Herdr payload parsing and command generation where practical.
6. Update README/cheatsheet with new live-control commands.

# Analysis: live-herdr-tmux-control

## Summary

Harden live-session control by treating Herdr as the preferred runtime backend and tmux as the fallback. Herdr exposes session/workspace/tab/pane/agent state through CLI JSON and a socket API; for this slice, use CLI wrappers where possible and keep raw socket integration for later.

## Herdr research notes

- Herdr uses a server/client model. Detach/reattach keeps server-owned processes alive.
- The hierarchy is sessions, workspaces, tabs, panes, and agents.
- Herdr CLI talks to the same local socket API as integrations and agents.
- CLI wrappers are recommended before raw socket API for scripts and simple orchestration.
- `session.snapshot` is the raw socket method for one-time local runtime cache bootstrap.
- `agent.list`, `agent.get`, `pane.list`, and `pane.get` expose live state; pane/agent records may include `agent_session` when integrations report native session identity.
- `agent prompt` is the preferred high-level prompt/send operation; it can optionally wait for state transitions.
- `pane read` / `agent read` can read recent output for summaries and status inspection.
- Native session restore depends on current Herdr integrations reporting session references.
- Supported restore commands include `claude --resume <id>`, `copilot --resume=<id>`, `hermes --resume <id>`, and others.

## Current gaps

- Herdr scanner only parses a shallow `agent list --json` payload and does not handle nested `result` payloads.
- `tss send` uses an outdated `herdr agent send` shape instead of `agent prompt`.
- There is no `tss run <session> -- <command>` to open a pane/window and run a command in the session directory.
- There is no `tss read <session>` to inspect recent live output.
- tmux supports command execution/read primitives but the CLI does not expose them yet.

## Risks

- Herdr is not installed in the current environment, so implementation must be validated with command generation and fake/shape-oriented tests until a real Herdr install is available.
- Herdr JSON response shapes can evolve; parsing should be permissive and tolerate nested `result` wrappers.
- Commands that create panes should support `--print` before execution.

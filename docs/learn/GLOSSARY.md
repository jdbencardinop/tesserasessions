# tss operator glossary

Canonical language for the inventory and runtime model. Lessons and runbooks
use these terms consistently.

## Inventory

**Source**:

A system that supplies session or runtime observations to `tss`, such as
Claude, Codex, Herdr, or tmux.

**Historical session**:

An independently resumable record discovered in an agent's persistent store.
It can exist without a live process.

**Inventory**:

The local SQLite projection maintained by explicit `tss scan` operations. It is
not the source tool's database and is not a live-process registry.

**Native ID**:

The source tool's stable resume identity. `tss` derives its own source-qualified
ID from the source and native ID.

**Authoritative snapshot**:

A successful complete source scan whose result can safely replace that source's
prior inventory rows.

**Incomplete snapshot**:

A scan that found valid rows but could not read every authoritative item.
Valid rows may refresh, but prior rows are not pruned.

## Live state

**Runtime instance**:

A currently observable agent process, pane, or surface reported by a live
provider. A runtime may match a historical session, but the concepts are not
interchangeable.

**Provider**:

A bounded live observer, currently Herdr or tmux, that reports availability and
runtime evidence.

**Runtime presence**:

Whether a matching runtime is `present`, `absent`, `stale`, or `unknown`.
Presence says nothing by itself about what the agent is doing.

**Agent state**:

The semantic state of a live agent: `working`, `ready`, `blocked`, `done`, or
`unknown`.

**Ready**:

A live agent can accept input. It is not the same as an inventory or dashboard
item with no session.

**Idle**:

A normalized inventory status for work not currently active. The runtime
contract deliberately uses `ready`, not `idle`, for a live waiting agent.

**Stale**:

Evidence exists but is older than its freshness contract or has been manually
classified as stale.

**Needs attention**:

A normalized inventory status indicating a blocker, failure, stale condition,
or operator follow-up. It is a rollup label, not a live-provider agent state.

## Returning to work

**Resume**:

Start the source tool using a historical session's exact native identity and
source-store context.

**Attach**:

Return to an existing matching live runtime when possible, otherwise use the
session's native resume command.

**Open**:

Create a new Herdr workspace or tmux session in the same project directory. It
does not resume the historical agent conversation.

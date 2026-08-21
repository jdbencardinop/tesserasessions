# Implementation Record: live-herdr-tmux-control

**Recorded**: 2026-07-22T14:33:20Z
**Files changed**: 7
**Patch size**: 17152 bytes
**Capture mode**: working-tree-all
**Pathspecs**: internal/adapters/herdr.go,internal/adapters/herdr_test.go,internal/cli/root.go,internal/cli/live_commands_test.go,README.md,docs/cheatsheet.md,docs/roadmap.md

## Change Summary

```
 README.md                  |   8 ++-
 docs/cheatsheet.md         |   4 ++
 docs/roadmap.md            |   5 +-
 internal/adapters/herdr.go | 109 ++++++++++++++++++++++++++++---------
 internal/cli/root.go       | 131 ++++++++++++++++++++++++++++++++++++++++++++-
 5 files changed, 227 insertions(+), 30 deletions(-)
```

## Capture Provenance

- **capture_mode**: `working-tree-all`
- **pathspecs**: internal/adapters/herdr.go, internal/adapters/herdr_test.go, internal/cli/root.go, internal/cli/live_commands_test.go, README.md, docs/cheatsheet.md, docs/roadmap.md
- **claim_ids**: (none)
- **base_commit**: `1ec3777112fee1e6caa4b72d24653f2ecdaedf90`
- **upper_commit**: `working-tree`

## Replay Instructions

To re-apply this feature to a clean checkout:

```bash
# From the feature's artifacts directory:
git apply .tpatch/features/live-herdr-tmux-control/artifacts/post-apply.patch
```

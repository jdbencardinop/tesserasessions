# Implementation Record: runtime-status-contract

**Recorded**: 2026-08-11T17:23:42Z
**Files changed**: 19
**Patch size**: 61782 bytes
**Capture mode**: working-tree-all

## Change Summary

```
 .tpatch/FEATURES.md                                |  1 +
 .../features/agent-source-expansion/status.json    |  4 ++
 CHANGELOG.md                                       |  5 +-
 README.md                                          | 29 ++++++++--
 docs/cheatsheet.md                                 | 43 ++++++++++++---
 docs/roadmap.md                                    | 57 ++++++++++++++++++--
 internal/adapters/herdr.go                         |  5 ++
 internal/adapters/scanner.go                       |  8 ++-
 internal/adapters/tmux.go                          |  6 +++
 internal/cli/root.go                               | 19 +++++--
 internal/core/model.go                             | 13 ++---
 internal/store/store.go                            | 61 ++++++++++++++++++++--
 internal/store/store_test.go                       | 48 +++++++++++++++++
 13 files changed, 266 insertions(+), 33 deletions(-)
```

## Capture Provenance

- **capture_mode**: `working-tree-all`
- **pathspecs**: (none)
- **claim_ids**: (none)
- **base_commit**: `3f5f340736cdf236e467f0134d80b416152da528`
- **upper_commit**: `working-tree`

## Replay Instructions

To re-apply this feature to a clean checkout:

```bash
# From the feature's artifacts directory:
git apply .tpatch/features/runtime-status-contract/artifacts/post-apply.patch
```


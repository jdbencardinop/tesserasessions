# Implementation Record: agent-source-expansion

**Recorded**: 2026-08-12T09:30:28Z
**Files changed**: 21
**Patch size**: 77483 bytes
**Capture mode**: working-tree-all

## Change Summary

```
 .tpatch/FEATURES.md                                |   2 +-
 .tpatch/features/agent-source-expansion/request.md |   2 +
 .../features/agent-source-expansion/status.json    |  24 +++-
 CHANGELOG.md                                       |   8 ++
 README.md                                          |   9 +-
 docs/cheatsheet.md                                 |  11 +-
 docs/roadmap.md                                    |  24 ++--
 internal/adapters/claude.go                        | 155 +++++++++++++++++----
 internal/adapters/claude_test.go                   |  96 +++++++++++--
 internal/adapters/copilot.go                       | 100 +++++++++----
 internal/adapters/scanner.go                       |  16 +++
 internal/cli/root.go                               |  26 +++-
 internal/cli/scan_runtime_test.go                  |  35 +++++
 internal/config/config.go                          |  58 ++++++++
 internal/core/model.go                             |  15 +-
 internal/store/store.go                            |  79 ++++++++++-
 internal/store/store_test.go                       |  92 ++++++++++--
 17 files changed, 637 insertions(+), 115 deletions(-)
```

## Capture Provenance

- **capture_mode**: `working-tree-all`
- **pathspecs**: (none)
- **claim_ids**: (none)
- **base_commit**: `b85896f709859575a1ba454dfb603b90183d7857`
- **upper_commit**: `working-tree`

## Replay Instructions

To re-apply this feature to a clean checkout:

```bash
# From the feature's artifacts directory:
git apply .tpatch/features/agent-source-expansion/artifacts/post-apply.patch
```

# Implementation Record: cli-inventory-core

**Recorded**: 2026-07-10T05:59:52Z
**Files changed**: 17
**Patch size**: 78901 bytes
**Capture mode**: working tree
**Pathspecs**: go.mod,go.sum,cmd/,internal/

## Change Summary

```
 internal/adapters/claude_test.go |  31 +++++++
 internal/core/model.go           | 179 +++++++++++++++++++++++++++++++++++++++
 2 files changed, 210 insertions(+)
```

## Replay Instructions

To re-apply this feature to a clean checkout:

```bash
# From the feature's artifacts directory:
git apply .tpatch/features/cli-inventory-core/artifacts/post-apply.patch
```


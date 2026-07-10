# Spec: smart-session-summaries

## Problem

Initial summaries use the first extracted text almost verbatim, which can be too noisy. The MVP needs better local titles and summaries while preserving the privacy default of no remote LLM calls.

## Acceptance criteria

1. `tss summarize <session>` derives a concise title from meaningful user/session text, not raw JSON or obvious tool noise.
2. The goal summary includes a goal and latest/next context when enough text exists.
3. Status can infer `needs_attention` and `done` from local text patterns.
4. Transcript/content reading remains lazy and read-only.
5. `go test ./...` passes.

## Out of scope

- Remote LLM summarization.
- Full semantic recap quality.
- Agent-specific perfect parsers for every transcript format.

## Plan

1. Improve candidate extraction and filtering.
2. Add title normalization helpers.
3. Improve summary composition.
4. Add tests for noisy text and title generation.

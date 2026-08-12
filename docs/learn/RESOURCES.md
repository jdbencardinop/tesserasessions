# tss operator resources

These are the trusted sources behind the learning path. Each entry states what
it establishes and when to use it.

## Knowledge

- [tss README](../../README.md)
  Product scope, installation, command overview, storage, and safety defaults.
  Use it for the supported public surface.
- [tss cheatsheet](../cheatsheet.md)
  Compact command syntax for daily work. Use it after completing the lessons.
- [Historical source adapters](../source-adapters.md)
  Store discovery, filtering, resume context, reconciliation, and privacy
  boundaries. Use it when a source is missing or its inventory looks wrong.
- [Runtime status provider contract](../runtime-status-contract.md)
  Normative ownership, freshness, matching, `runtime_presence`, and
  `agent_state` semantics. Use it when diagnosing live state or integrating a
  consumer such as `tws`.
- [`teach` skill by Matt Pocock](https://github.com/mattpocock/skills/tree/main/skills/productivity/teach)
  Method adapted for mission-driven lessons, retrieval practice, references,
  and evidence-based learning records. Use it when extending this curriculum.
  The source is available under its
  [MIT license](https://github.com/mattpocock/skills/blob/main/LICENSE).
- [GitHub Copilot CLI documentation](https://docs.github.com/en/copilot/how-tos/copilot-cli/use-copilot-cli/overview)
  Official Copilot CLI behavior and safety model. Use it for source-tool
  questions beyond the `tss` adapter boundary.
- [Claude Code documentation](https://code.claude.com/docs/en/overview)
  Official Claude Code surfaces and setup. Use it for Claude behavior rather
  than inferring from stored metadata.
- [OpenAI Codex CLI documentation](https://developers.openai.com/codex/cli/)
  Official Codex CLI usage and resume behavior. Use it for Codex behavior
  outside active-session inventory.
- [OpenCode documentation](https://opencode.ai/docs)
  Official OpenCode installation and usage. Use it for application behavior,
  not as a substitute for the adapter contract.
- [tmux wiki](https://github.com/tmux/tmux/wiki)
  Primary project guidance for tmux sessions and panes. Use it when a generated
  tmux command behaves differently from expected.

## Wisdom

- [tesserasessions issues](https://github.com/jdbencardinop/tesserasessions/issues)
  Report reproducible inventory, matching, resume, documentation, or privacy
  defects with redacted output.
- Official issue trackers linked from each source tool's documentation
  Use them when the behavior reproduces in the source tool without `tss`.

Community advice can help diagnose environment-specific behavior, but it does
not override the contracts above.

## Gaps

- Hermes Agent does not yet have a pinned, operator-facing persistence contract
  in this curriculum. Use the `tss` source adapter guide for supported behavior.
- Herdr's live JSON/status contract should be linked here once a stable official
  reference is available.
- t3code is not yet a supported historical source.

# Spec: operator-learning-runbook

## Problem

The command reference is accurate but assumes the reader already understands
the `tss` state model. New operators need a short, practice-oriented path that
teaches what evidence exists, which command is safe at each step, and how to
return to work without conflating historical inventory with live runtime state.

## Acceptance criteria

1. Use tpatch Path B only. Author all phase artifacts and implementation
   directly without invoking the configured AI provider.
2. Create a self-contained learning workspace under `docs/learn/` with:
   - `README.md`;
   - `MISSION.md`;
   - `RESOURCES.md`;
   - `GLOSSARY.md`;
   - reusable local assets;
   - four numbered HTML lessons;
   - one printable first-day runbook;
   - local learning-record and notes templates.
3. Define a concise, observable operator mission. Success criteria must include
   building an inventory, explaining the evidence model, safely returning to a
   session, and diagnosing incomplete or conflicting state.
4. Make `docs/learn/README.md` the course index:
   - state prerequisites and a recommended lesson order;
   - distinguish lessons, references, and local learner state;
   - attribute the adapted teaching methodology and link its MIT license;
   - make clear that individual learning records are evidence of demonstrated
     understanding, not an activity log.
5. Curate `RESOURCES.md` using annotated entries:
   - prefer local contracts and official source documentation;
   - state what each source establishes and when to use it;
   - separate knowledge sources from community/wisdom sources;
   - list unresolved documentation gaps explicitly.
6. Establish canonical definitions for at least:
   - source;
   - historical session;
   - inventory;
   - runtime instance;
   - provider;
   - authoritative snapshot;
   - incomplete snapshot;
   - native ID;
   - resume, attach, and open;
   - runtime presence;
   - agent state;
   - ready, idle, stale, and needs attention.
7. Produce a printable first-day runbook that covers:
   - `tss doctor`;
   - `tss scan` and scan-result interpretation;
   - `tss list`, `search`, and `show`;
   - previewing `attach`/`open` with `--print`;
   - side-effect-free `status --json`;
   - curation commands;
   - a troubleshooting decision table;
   - privacy and side-effect labels.
8. Create four short lessons, each with one tangible outcome:
   1. build and assess an inventory;
   2. find and inspect the right session;
   3. choose and preview a safe return path;
   4. diagnose runtime presence separately from agent state.
9. Every lesson must:
   - use actual current `tss` command syntax;
   - use only synthetic paths, IDs, and sample output;
   - link to the mission, glossary, runbook, adjacent lessons, and one primary
     source;
   - include a retrieval or decision exercise with immediate feedback;
   - remain short enough to complete in roughly ten minutes.
10. Reuse local assets:
    - one stylesheet shared by every HTML document;
    - one progressively enhanced lesson script shared by lessons;
    - no inline reusable CSS/JavaScript duplication;
    - no CDN, analytics, network request, font download, or third-party script.
11. HTML is usable from a local checkout:
    - relative links resolve;
    - core content works with JavaScript disabled;
    - keyboard focus is visible;
    - narrow-screen, reduced-motion, high-contrast, and print styles exist;
    - interactive feedback is announced accessibly.
12. Teach safety precisely:
    - `status --json` is side-effect-free and does not scan;
    - `scan` writes only the local `tss` inventory and reads source stores;
    - `search --content` is explicit content access;
    - interactive/control commands are previewed with `--print` first;
    - users are never instructed to edit SQLite or source stores directly.
13. Keep learner-specific state out of Git:
    - ignore `docs/learn/NOTES.md`;
    - ignore `docs/learn/learning-records/`;
    - commit reusable templates only;
    - do not include personal claims, paths, session IDs, transcripts,
      screenshots, databases, or attachments.
14. Link the learning path from `README.md` and `docs/cheatsheet.md`. Mark the
    roadmap feature delivered only after implementation is complete.
15. Verify all taught commands and flags against current `bin/tss ... --help`,
    all internal relative links against the committed tree, and all HTML
    documents for balanced structure and local asset references.
16. Run repository tests, formatting/diff checks appropriate to the changed
    files, and independent documentation/usability review until approved.

## Out of scope

- Hosted documentation or a documentation deployment pipeline.
- Telemetry, learner analytics, accounts, or synchronized progress.
- Tutorials for operating Claude, Copilot, Hermes, OpenCode, Codex, Herdr, or
  tmux themselves.
- Reading or bundling real agent transcripts.
- Personalized adaptive lesson generation.
- Product CLI or database behavior changes.

## Implementation plan

1. Inventory current command help, contracts, and official source material.
2. Write the mission, resource map, glossary, templates, and course index.
3. Build accessible shared assets and the printable first-day runbook.
4. Build four focused lessons using the shared assets and synthetic examples.
5. Wire discovery links, ignore local state, and update roadmap/changelog.
6. Validate commands, links, HTML, privacy, and accessibility assumptions.
7. Run independent review, revise, record, verify, land, and push.

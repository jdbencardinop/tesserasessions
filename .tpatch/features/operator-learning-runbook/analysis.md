# Analysis: operator-learning-runbook

## Summary

Add a repository-owned operator learning workspace under `docs/learn/`.
Existing README and cheatsheet material explains commands, but it does not
provide a progressive path from first scan to safe resume and troubleshooting.

The workspace adapts the learning model from Matt Pocock's MIT-licensed
[`teach` skill](https://github.com/mattpocock/skills/tree/main/skills/productivity/teach):
ground lessons in a concrete mission, use trusted sources for knowledge, build
skills through short practice and feedback, preserve reusable references, and
record demonstrated learning separately from material merely covered.

This is an adaptation, not a copy of the source workspace. All prose, examples,
HTML, CSS, and JavaScript will be original to `tss`, with attribution to the
methodology.

## User problem

`tss` now spans historical stores, live runtime observation, session curation,
native resume, Herdr/tmux control, and a machine-readable status contract. A
new operator can find individual commands in the README but still lacks a
coherent answer to:

1. What should I run first?
2. Which state belongs to the source tool, the inventory, or a live runtime?
3. How do I inspect and resume without accidentally executing first?
4. How do I distinguish missing, skipped, incomplete, stale, and blocked state?
5. Which commands are metadata-only, content-reading, side-effect-free, or
   interactive?

## Teaching model

### Mission

The generic operator mission is concrete: independently locate the right agent
work, judge the evidence available, and safely return to or troubleshoot it
without exposing transcript content or confusing inventory state with runtime
state.

### Knowledge

Claims should come from:

- current `tss` command help and repository contract documentation;
- official documentation for supported source tools and live backends;
- the referenced teaching methodology for lesson design.

Every external resource must be annotated with what it establishes and when to
use it. Missing primary documentation should be listed as a gap rather than
filled with an unsupported claim.

### Skills

Short lessons should build one observable capability at a time:

1. establish source and inventory health;
2. find and inspect a session;
3. preview and choose a safe return path;
4. diagnose live runtime presence separately from agent state;
5. curate the inventory for the next handoff.

Exercises use real `tss` commands, synthetic paths and IDs, `--print` before
interactive actions, and immediate self-check feedback. Retrieval prompts
should test decisions rather than recognition alone.

### Wisdom

The curriculum can identify high-trust issue trackers and communities for
source-specific behavior, but it must not present community advice as a
contract. The `tss` issue tracker is the escalation path for reproducible
inventory or control defects.

## Workspace shape

Committed, reusable material:

```text
docs/learn/
├── README.md
├── MISSION.md
├── RESOURCES.md
├── GLOSSARY.md
├── assets/
│   ├── course.css
│   └── lesson.js
├── lessons/
│   ├── 0001-build-an-inventory.html
│   ├── 0002-find-the-right-session.html
│   ├── 0003-return-safely.html
│   └── 0004-diagnose-live-state.html
├── reference/
│   └── first-day-runbook.html
└── templates/
    ├── learning-record.md
    └── notes.md
```

Local learner state:

```text
docs/learn/NOTES.md
docs/learn/learning-records/
```

The local paths must be gitignored. The repository commits formats and generic
course material, never claims about a particular user's knowledge.

## Integration points

- `README.md`: add one discoverable learning-path link.
- `docs/cheatsheet.md`: link the runbook for first-time operators.
- `docs/roadmap.md`: mark the feature delivered only after landing.
- `.gitignore`: protect local notes and learning records.
- No CLI, database, adapter, or schema changes are required.

## Safety and privacy

- Use dummy IDs and paths; never capture a real inventory, username, home
  directory, transcript, screenshot, or database.
- Clearly label `status --json` as side-effect-free and `scan` as a local
  inventory write.
- Clearly label `search --content` as opt-in content access.
- Teach `attach`, `open`, `send`, `read`, and `run` with `--print` first.
- Do not tell users to edit the inventory database or source stores directly.
- Keep lesson assets local; no analytics, CDNs, external scripts, or network
  calls.

## Compatibility

The change is additive documentation. Static lessons link to current public
commands and must remain usable from a local checkout without a web server.
Relative links, keyboard access, reduced-motion preferences, narrow screens,
and print output need explicit validation.

## Risks

- Command drift can make lessons actively misleading. Acceptance checks should
  compare every taught command and flag to current help.
- A broad "course" can overwhelm new users. Each lesson must stay short and
  produce one tangible win.
- Committing learner records would conflate product documentation with personal
  state. Ignore local records and provide templates only.
- Copying the upstream skill's exact templates would create unnecessary
  licensing and maintenance coupling. Preserve the method, cite it, and author
  `tss`-specific content.

## Recommendation

Proceed through Path B. Build the first-day journey and four lessons as one
documentation slice, review it for command accuracy and operator usability, and
defer hosted documentation, telemetry, source-tool tutorials, and personalized
adaptive teaching automation.

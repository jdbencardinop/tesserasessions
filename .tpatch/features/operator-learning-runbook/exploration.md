# Exploration: operator-learning-runbook

## Existing documentation surfaces

- `README.md`
  - installation and quick-start sequence;
  - command table;
  - architecture, source adapters, privacy, and tpatch workflow.
- `docs/cheatsheet.md`
  - concise daily command reference;
  - source list, status request example, and development checks.
- `docs/runtime-status-contract.md`
  - authoritative ownership, side-effect, matching, freshness, and aggregate
    semantics for `status --json`.
- `docs/source-adapters.md`
  - authoritative source-store locations, filtering, resume context, and
    metadata privacy boundary.
- `docs/roadmap.md`
  - feature DAG, delivered slices, and the requested learning milestone.

The new material should link to these documents rather than duplicate their
complete contracts.

## Current command surface

The installed `bin/tss` exposes:

- health and ingestion: `doctor`, `scan [--source] [--json]`;
- discovery: `list`, `show`, `search`;
- curation: `summarize`, `title`, `mark`, `pin`, `tag`;
- return/control: `attach`, `open`, `send`, `read`, `run`;
- live contract: `status --json`;
- shell integration: `completion`.

`list`, `show`, `search`, and `scan` support JSON where appropriate. Control
commands support `--print`; `search --content` is the only documented
content-reading search mode.

## Minimal change set

### New generic learning material

- `docs/learn/README.md`
- `docs/learn/MISSION.md`
- `docs/learn/RESOURCES.md`
- `docs/learn/GLOSSARY.md`
- `docs/learn/assets/course.css`
- `docs/learn/assets/lesson.js`
- `docs/learn/reference/first-day-runbook.html`
- `docs/learn/lessons/0001-build-an-inventory.html`
- `docs/learn/lessons/0002-find-the-right-session.html`
- `docs/learn/lessons/0003-return-safely.html`
- `docs/learn/lessons/0004-diagnose-live-state.html`
- `docs/learn/templates/learning-record.md`
- `docs/learn/templates/notes.md`

### Existing files

- `.gitignore`
  - ignore `docs/learn/NOTES.md`;
  - ignore `docs/learn/learning-records/`.
- `README.md`
  - link the learning path after quick start.
- `docs/cheatsheet.md`
  - link the first-day runbook before the daily loop.
- `docs/roadmap.md`
  - move the learning feature from Next to Delivered and update its table row.
- `CHANGELOG.md`
  - record the operator learning workspace.

No Go source, SQLite schema, adapter, or command changes are needed.

## HTML design

Every HTML document uses semantic landmarks:

```html
<header>...</header>
<nav aria-label="Course">...</nav>
<main>...</main>
<footer>...</footer>
```

Lessons share:

- skip link;
- breadcrumb/course navigation;
- outcome and estimated-time callout;
- compact concept section;
- terminal exercise using synthetic values;
- decision exercise;
- native `<details>` answer fallback;
- progressively enhanced quiz feedback with `aria-live`;
- previous/next and reference links.

`assets/lesson.js` should only enhance forms marked with `data-quiz`. It must not
send data, store progress, or hide the native answer fallback.

`assets/course.css` supplies system-font typography, reusable cards and
terminal blocks, visible focus, responsive layout, high-contrast-compatible
colors, reduced-motion behavior, and print rules. HTML files contain no inline
style or reusable script.

## Canonical lesson examples

All examples use neutral data such as:

- `/workspaces/demo-api`;
- `/repos/demo`;
- `claude-demo123`;
- `feature/demo`;
- `{"query_id":"demo-api", ...}`.

No generated example may include the current username, home directory, Git
remote, inventory database, session metadata, or transcript text.

## Resource provenance

Primary sources:

- repository README, source adapter guide, and runtime contract;
- GitHub Copilot CLI official documentation;
- Claude Code official documentation;
- OpenAI Codex CLI official documentation;
- OpenCode official documentation;
- tmux project documentation.

Method source:

- Matt Pocock's `teach` skill and format notes, MIT licensed.

The resource list should annotate uses and explicitly record gaps for Hermes
and Herdr operator-facing session metadata contracts instead of inferring them.

## Local learner state

The committed `templates/learning-record.md` should be intentionally minimal:
what is now understood, evidence, and why it changes the next lesson. It must
warn that coverage is not evidence.

The committed `templates/notes.md` is a copyable scratch format. Operators may
copy it to ignored `docs/learn/NOTES.md`. Learning records go into ignored
`docs/learn/learning-records/NNNN-slug.md`.

## Validation

```sh
# Existing project gates
make test
make lint
make build

# Command contract checks
bin/tss --help
bin/tss doctor --help
bin/tss scan --help
bin/tss list --help
bin/tss show --help
bin/tss search --help
bin/tss attach --help
bin/tss open --help
bin/tss status --help

# Repository integrity
git diff --check
tpatch feature deps --validate-all
```

Additionally inspect every relative `href`/`src` target, confirm no external
lesson assets or real local identifiers, parse each HTML file with a standard
library parser, and perform independent command-accuracy/usability review.

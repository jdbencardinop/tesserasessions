# Learn to operate tss

This path teaches one practical outcome: find the right agent work, understand
the evidence available, and return to it safely.

**Prerequisites:** build or install `tss`, and have at least one supported
source configured. Start with `tss doctor`; you do not need to expose or commit
any transcript content.

## Learning path

| Lesson | Tangible win | Time |
| --- | --- | --- |
| [1. Build an inventory](lessons/0001-build-an-inventory.html) | Judge source and scan health without guessing. | 8 min |
| [2. Find the right session](lessons/0002-find-the-right-session.html) | Narrow metadata and inspect one exact session. | 8 min |
| [3. Return safely](lessons/0003-return-safely.html) | Choose attach, resume, or open and preview the command. | 10 min |
| [4. Diagnose live state](lessons/0004-diagnose-live-state.html) | Separate runtime presence from agent state. | 10 min |

Keep the printable
[first-day runbook](reference/first-day-runbook.html) nearby after the lessons.
Use the [glossary](GLOSSARY.md) when a state term is ambiguous and the
[resource map](RESOURCES.md) when a claim needs a primary source.

The lessons are standalone local HTML files. Open them directly in a browser;
no server, CDN, analytics, network request, or account is required.

## References are not lessons

Lessons build a decision skill through short practice. References compress the
answer for later use. Reading the runbook is useful, but retrieval and correct
application are stronger evidence that the concept will remain available under
pressure.

## Keep your learning state local

The repository commits reusable material, not claims about an individual.
Personal notes and learning records are ignored by Git.

```sh
cp docs/learn/templates/notes.md docs/learn/NOTES.md
mkdir -p docs/learn/learning-records
cp docs/learn/templates/learning-record.md \
  docs/learn/learning-records/0001-first-inventory.md
```

Write a learning record only after you can demonstrate a non-obvious decision,
correct a misconception, or establish meaningful prior knowledge. It is not a
session log.

## Method attribution

The learning architecture adapts Matt Pocock's MIT-licensed
[`teach` skill](https://github.com/mattpocock/skills/tree/main/skills/productivity/teach):
mission grounding, trusted knowledge, short skill practice, reusable
references, and evidence-based learning records. All `tss` lesson content and
assets in this directory are original to this project.

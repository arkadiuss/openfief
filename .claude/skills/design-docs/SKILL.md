---
name: design-docs
description: Edit OpenFief's design documents (docs/DESIGN.md, docs/proposals/, docs/research.md, docs/glossary.md) without breaking their structure or IDs. Use before any change to those files.
---

# Editing the design documents

## Where things go

| File | Holds |
|---|---|
| `docs/DESIGN.md` | The current, decided design: ground rules (GR), decisions (D), and the tensions (T) and open questions (OQ) that block M0. A snapshot: no dates, no history, no status notes. |
| `docs/proposals/<area>.md` | Undecided proposals (P), with that area's tensions and open questions. |
| `docs/research.md` | External facts (R), each with its date and sources. |
| `docs/glossary.md` | Domain terms (the ubiquitous language) and other jargon. |

Development process belongs in `docs/dev/`, never in the design documents.

## Rules

- **Ground rules** are never edited, reworded, merged or reinterpreted. A conflict becomes a tension.
- **Decisions come only from the author.** When the author decides a proposal, write it into DESIGN.md as a decision and delete it from its proposal file. When an open question is answered or a tension resolved, delete it. Git keeps the history, so never leave status notes, dates or "revised" markers.
- **DESIGN.md states facts plainly.** It never links to external research; a fact a decision depends on is one plain sentence where it's used. Proposals may cite research by ID.
- **IDs are stable and never reused.** A new item takes the next free number for its tag.
- **Explain every abbreviation** on first use, or add it to the glossary.

## Checks after editing

Run `task docs:check`; the pre-commit hook runs it too. It fails on any referenced tag that isn't defined, and on any changed ground-rule line. It also fails when an ID is defined twice. If you touched a Mermaid block, check that it renders.

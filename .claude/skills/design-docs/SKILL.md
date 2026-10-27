---
name: design-docs
description: Edit OpenFief's docs/DESIGN.md or docs/SPEC.md without breaking provenance tags (GR, D, R, P, P-CC, T, OQ). Use before any change to those two files.
---

# Editing the design records

DESIGN.md §1 holds the tag legend. These rules keep provenance intact.

## Rules

- **Ground rules `[GR-n]`** are never edited, reworded, merged or reinterpreted. A conflict becomes a new tension `[T-n]`, tagged with an `[OQ-n]`.
- **Decisions `[D-n]`** come only from the author, in their own words or by picking an option. Record each in DESIGN.md §10.1 with a note when it went against a recommendation or was the author's own answer. Revise a decision in place with a "revised" note; never silently rewrite it.
- **Your additions** are `[P-CC-n]` (proposal) or `[OQ-n]` (open question), using the next free number. Never present a proposal or a research finding as decided.
- **Research findings `[R-n]`** carry their date, the label *"As of <date>, re-verify before relying"*, and source URLs.
- **When something gets settled,** add an italic status note after the original item (`*Status <date>: accepted by D-n*`, `*Answered by D-n*`, `*Withdrawn*`). Keep the original text; never delete history.
- **SPEC.md requirements:**
  - MUST only when the requirement traces to a GR.
  - SHOULD plus *(pending author approval)* when it traces only to proposals.
  - *(decided: D-n)* when a decision settles it.
  - SPEC.md §2 is frozen (D-33): new requirements go to kibi (docs/dev/kibi.md).
- **Explain every abbreviation on first use**, and add new jargon to the DESIGN.md glossary.

## Checks after editing

Run `task docs:check`; the pre-commit hook runs it too. It fails on any referenced tag that isn't defined, and on any changed ground-rule line (compared against `HEAD`, or against `DOCS_BASE` in CI). If you touched a Mermaid block, check it in GitHub's preview or with mermaid-cli.

---
title: The docs check must fail when a design document references a provenance tag that is not defined.
status: open
text_ref: .claude/skills/design-docs/SKILL.md
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:claude-skills-design-docs-skill-md
  - lane:strict
semantic_text: The docs check must fail when a design document references a provenance tag that is not defined.
logic_claims:
  - CLAIM-DD2B1D93A96EADE4
semantic_clauses:
  - The docs check must fail when a design document references a provenance tag that is not defined.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: eee11227775c959aca21ad442dbd4f2e70cb314c389925d5c61b74cf6202f386
semantic_inventory:
  - claim_key: CLAIM-DD2B1D93A96EADE4
    claim_text: The docs check must fail when a design document references a provenance tag that is not defined
    role: normative
    status: modeled
    span:
      start: 0
      end: 95
origin:
  kind: agent
  recorded_at: '2026-10-27T20:32:02.429Z'
id: REQ-tooling-docs-undefined-tag
type: req
---
The docs check must fail when a design document references a provenance tag that is not defined.

## Context

Keeps the design documents' IDs consistent: every GR, D, P, T, OQ and R reference must point at a defined item. Recorded by Claude Code while wiring the tooling; the author made the Taskfile the single entry point for all tooling, and this check runs as task docs:check and in the pre-commit hook.

## Source

> It fails on any referenced tag that isn't defined, and on any changed ground-rule line.

.claude/skills/design-docs/SKILL.md#checks-after-editing

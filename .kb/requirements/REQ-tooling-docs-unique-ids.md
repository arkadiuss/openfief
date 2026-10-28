---
title: The docs check must fail when an ID is defined more than once across the design documents.
status: open
text_ref: .claude/skills/design-docs/SKILL.md
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:claude-skills-design-docs-skill-md
  - lane:strict
semantic_text: The docs check must fail when an ID is defined more than once across the design documents.
logic_claims:
  - CLAIM-824515FA892E178F
semantic_clauses:
  - The docs check must fail when an ID is defined more than once across the design documents.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 88aee7f35c9e958c1053942fd29e9157c9ab77da498d1ea3e7a5b9e3c4dec40f
semantic_inventory:
  - claim_key: CLAIM-824515FA892E178F
    claim_text: The docs check must fail when an ID is defined more than once across the design documents
    role: normative
    status: modeled
    span:
      start: 0
      end: 89
origin:
  kind: agent
  recorded_at: '2026-10-28T03:53:22.880Z'
id: REQ-tooling-docs-unique-ids
type: req
---
The docs check must fail when an ID is defined more than once across the design documents.

## Context

With the design split across DESIGN.md, the proposal files, research and the glossary, two files could define the same ID; references would then be ambiguous. Recorded by Claude Code while wiring the tooling; the author made the Taskfile the single entry point for all tooling, and this check runs as task docs:check and in the pre-commit hook.

## Source

> It also fails when an ID is defined twice.

.claude/skills/design-docs/SKILL.md#checks-after-editing

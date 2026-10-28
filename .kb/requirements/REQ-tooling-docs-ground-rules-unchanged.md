---
title: The docs check must fail when a ground-rule line in docs/DESIGN.md changes.
status: open
text_ref: .claude/skills/design-docs/SKILL.md
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:claude-skills-design-docs-skill-md
  - lane:strict
semantic_text: The docs check must fail when a ground-rule line in docs/DESIGN.md changes.
logic_claims:
  - CLAIM-2562BA4C2FB51222
semantic_clauses:
  - The docs check must fail when a ground-rule line in docs/DESIGN.md changes.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: b47544c1f650b8d043062b45d6fcbc109a0f601c55fe5ac4589c375fad8812b0
semantic_inventory:
  - claim_key: CLAIM-2562BA4C2FB51222
    claim_text: The docs check must fail when a ground-rule line in docs/DESIGN.md changes
    role: normative
    status: modeled
    span:
      start: 0
      end: 74
origin:
  kind: agent
  recorded_at: '2026-10-27T20:32:16.797Z'
id: REQ-tooling-docs-ground-rules-unchanged
type: req
---
The docs check must fail when a ground-rule line in docs/DESIGN.md changes.

## Context

Ground rules are written by the author and never edited (docs/DESIGN.md, Ground rules). Recorded by Claude Code while wiring the tooling; the author made the Taskfile the single entry point for all tooling, and this check runs as task docs:check and in the pre-commit hook.

## Source

> It fails on any referenced tag that isn't defined, and on any changed ground-rule line.

.claude/skills/design-docs/SKILL.md#checks-after-editing

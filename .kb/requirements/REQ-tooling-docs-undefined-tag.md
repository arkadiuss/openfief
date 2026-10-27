---
title: The docs check must fail when docs/DESIGN.md or docs/SPEC.md reference a provenance tag that is not defined.
status: open
text_ref: .claude/skills/design-docs/SKILL.md
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:claude-skills-design-docs-skill-md
  - lane:strict
semantic_text: The docs check must fail when docs/DESIGN.md or docs/SPEC.md reference a provenance tag that is not defined.
logic_claims:
  - CLAIM-F52906E2A6B6960A
semantic_clauses:
  - The docs check must fail when docs/DESIGN.md or docs/SPEC.md reference a provenance tag that is not defined.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: adba5d228dc8c0c3310ed8477cf3c8c965607fd1b6caca42e1c7d2195e0a67c3
semantic_inventory:
  - claim_key: CLAIM-F52906E2A6B6960A
    claim_text: The docs check must fail when docs/DESIGN.md or docs/SPEC.md reference a provenance tag that is not defined
    role: normative
    status: modeled
    span:
      start: 0
      end: 107
origin:
  kind: agent
  recorded_at: '2026-10-27T20:32:02.429Z'
id: REQ-tooling-docs-undefined-tag
type: req
---
The docs check must fail when docs/DESIGN.md or docs/SPEC.md reference a provenance tag that is not defined.

## Context

Keeps the design records' provenance intact: every GR, D, AC, R, P, P-CC, T and OQ reference must point at a defined item (DESIGN.md section 1, tag legend). Recorded by Claude Code while wiring the tooling under D-41; it runs as task docs:check and in the pre-commit hook.

## Source

> It fails on any referenced tag that isn't defined, and on any changed ground-rule line

.claude/skills/design-docs/SKILL.md#checks-after-editing

---
title: The read API's yearly tax must say when it is an estimate over the days so far, its settings must list each tax wrapper's account forms and whether it takes earlier payments, its accounts must carry their form and earlier payments, and the built-in UI must mark an estimate.
status: open
text_ref: docs/DESIGN.md#D-146
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-146
  - lane:strict
semantic_text: The read API's yearly tax must say when it is an estimate over the days so far, its settings must list each tax wrapper's account forms and whether it takes earlier payments, its accounts must carry their form and earlier payments, and the built-in UI must mark an estimate.
logic_claims:
  - CLAIM-682E30E35B797AD3
semantic_clauses:
  - The read API's yearly tax must say when it is an estimate over the days so far, its settings must list each tax wrapper's account forms and whether it takes earlier payments, its accounts must carry their form and earlier payments, and the built-in UI must mark an estimate.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: dda98fc3311883c2f744db67a62c276004c0bcf04dfcd127b6c9178376518204
semantic_inventory:
  - claim_key: CLAIM-682E30E35B797AD3
    claim_text: The read API's yearly tax must say when it is an estimate over the days so far, its settings must list each tax wrapper's account forms and whether it takes earlier payments, its accounts must carry their form and earlier payments, and the built-in UI must mark an estimate
    role: normative
    status: modeled
    span:
      start: 0
      end: 273
origin:
  kind: agent
  recorded_at: '2026-11-01T20:13:40.084Z'
id: REQ-views-yearly-tax-estimate-and-terms
type: req
---
The read API's yearly tax must say when it is an estimate over the days so far, its settings must list each tax wrapper's account forms and whether it takes earlier payments, its accounts must carry their form and earlier payments, and the built-in UI must mark an estimate.

## Context

D-146 makes the current year's OKI tax an estimate over the days so far; D-147 and D-149 add what the owner enters for an account, which readers need to show and offer. Modeled by Claude Code for the taxes package D-144..D-150; the author has not reviewed it yet.

## Source

> **[D-146]** An OKI account counts as held from its first transaction, and the current year's tax is an estimate over the days so far.

docs/DESIGN.md#D-146

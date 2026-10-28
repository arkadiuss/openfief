---
title: 'Each input change must enqueue revaluations by fixed rules: a price, the accounts that held the instrument from its date; an FX rate or index reading, every account from its date or its month''s first day; bond terms, their holders from their first transaction; a tax wrapper, its account from its first transaction; an import, each changed account from its earliest affected date; the end of day, every account from its valued-through date.'
status: open
text_ref: docs/DESIGN.md#D-36
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-36
  - lane:strict
semantic_text: 'Each input change must enqueue revaluations by fixed rules: a price, the accounts that held the instrument from its date; an FX rate or index reading, every account from its date or its month''s first day; bond terms, their holders from their first transaction; a tax wrapper, its account from its first transaction; an import, each changed account from its earliest affected date; the end of day, every account from its valued-through date.'
logic_claims:
  - CLAIM-027C2A8BCB07FB43
semantic_clauses:
  - 'Each input change must enqueue revaluations by fixed rules: a price, the accounts that held the instrument from its date; an FX rate or index reading, every account from its date or its month''s first day; bond terms, their holders from their first transaction; a tax wrapper, its account from its first transaction; an import, each changed account from its earliest affected date; the end of day, every account from its valued-through date.'
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 25a20759626654b605beac89304a1a9cb91d09cb170dacec250fa6c99a836378
semantic_inventory:
  - claim_key: CLAIM-027C2A8BCB07FB43
    claim_text: 'Each input change must enqueue revaluations by fixed rules: a price, the accounts that held the instrument from its date; an FX rate or index reading, every account from its date or its month''s first day; bond terms, their holders from their first transaction; a tax wrapper, its account from its first transaction; an import, each changed account from its earliest affected date; the end of day, every account from its valued-through date'
    role: normative
    status: modeled
    span:
      start: 0
      end: 439
origin:
  kind: agent
  recorded_at: '2026-10-28T07:30:32.396Z'
id: REQ-valuation-revaluation-triggers
type: req
---
Each input change must enqueue revaluations by fixed rules: a price, the accounts that held the instrument from its date; an FX rate or index reading, every account from its date or its month's first day; bond terms, their holders from their first transaction; a tax wrapper, its account from its first transaction; an import, each changed account from its earliest affected date; the end of day, every account from its valued-through date.

## Context

D-36 makes any change to the inputs enqueue a revaluation and makes what triggers one not configurable. The rules are the simple ones the orchestrator asked for: FX rates and index readings revalue every account. Refines REQ-decision-d36-10-fixed-revaluation-triggers. Modeled by Claude Code for the M0 revaluation task (D-36); the author has not reviewed it yet.

## Source

> What triggers a revaluation is not configurable.

docs/DESIGN.md D-36

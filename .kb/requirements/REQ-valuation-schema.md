---
title: The valuation context must keep its inputs, holding valuations, valued-through dates and pending revaluations in its own valuation schema with exact numerics, created and upgraded by its own migrations, and River's tables in a river schema.
status: open
text_ref: docs/DESIGN.md#D-36
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-36
  - lane:strict
semantic_text: The valuation context must keep its inputs, holding valuations, valued-through dates and pending revaluations in its own valuation schema with exact numerics, created and upgraded by its own migrations, and River's tables in a river schema.
logic_claims:
  - CLAIM-606351FD31E9390C
semantic_clauses:
  - The valuation context must keep its inputs, holding valuations, valued-through dates and pending revaluations in its own valuation schema with exact numerics, created and upgraded by its own migrations, and River's tables in a river schema.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 30a64cce2c9a9e4bc742bd04b22183539e3f3047418c7fbd8f71da667dbac692
semantic_inventory:
  - claim_key: CLAIM-606351FD31E9390C
    claim_text: The valuation context must keep its inputs, holding valuations, valued-through dates and pending revaluations in its own valuation schema with exact numerics, created and upgraded by its own migrations, and River's tables in a river schema
    role: normative
    status: modeled
    span:
      start: 0
      end: 239
origin:
  kind: agent
  recorded_at: '2026-10-28T07:59:03.178Z'
id: REQ-valuation-schema
type: req
---
The valuation context must keep its inputs, holding valuations, valued-through dates and pending revaluations in its own valuation schema with exact numerics, created and upgraded by its own migrations, and River's tables in a river schema.

## Context

Each context owns its schema (docs/dev/code-organization.md, proposal) and money is never floating point; stored rows must rebuild exactly (D-36). The runner copies the ledger store's so the contexts share no adapter code. Modeled by Claude Code for the M0 revaluation task (D-36); the author has not reviewed it yet.

## Source

> Valuations are stored rows derived from the transaction log, prices, FX rates and index readings, and can be rebuilt from them at any time.

docs/DESIGN.md D-36

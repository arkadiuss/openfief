---
title: The holding of an instrument in an account on a date must be the sum of the units in the holding postings of the account's current transactions dated on or before that date.
status: open
text_ref: docs/DESIGN.md#D-18
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-18
  - lane:strict
semantic_text: The holding of an instrument in an account on a date must be the sum of the units in the holding postings of the account's current transactions dated on or before that date.
logic_claims:
  - CLAIM-DD953DC648DF42E7
semantic_clauses:
  - The holding of an instrument in an account on a date must be the sum of the units in the holding postings of the account's current transactions dated on or before that date.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 376cf8d1ad1fb33263fffe9280ac2a5646f6f7cfe9d42dee48a4fe01492a5054
semantic_inventory:
  - claim_key: CLAIM-DD953DC648DF42E7
    claim_text: The holding of an instrument in an account on a date must be the sum of the units in the holding postings of the account's current transactions dated on or before that date
    role: normative
    status: modeled
    span:
      start: 0
      end: 172
origin:
  kind: agent
  recorded_at: '2026-10-28T05:39:14.275Z'
id: REQ-ledger-holding-derivation
type: req
---
The holding of an instrument in an account on a date must be the sum of the units in the holding postings of the account's current transactions dated on or before that date.

## Context

Refines the firm requirement REQ-AUTO-BE6EA3B9F864EA40 (D-18). The glossary defines a holding as the quantity of one instrument in one account on a date, derived from postings. A date includes its own transactions (end of day), matching how brokers report positions and how balance assertions are checked. Modeled by Claude Code for the M0 ledger task; the author has not reviewed it yet.

## Source

> Holdings and tax lots are derived from it.

docs/DESIGN.md D-18

> The quantity of one instrument in one account on a date, derived from postings.

docs/glossary.md (Holding)

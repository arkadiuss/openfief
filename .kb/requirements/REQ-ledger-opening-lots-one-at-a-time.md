---
title: Opening lots added one at a time to an account must all be dated on its go-live date when it is seeded, on one date when it is not, and be refused for an account not seeded that has a current transaction on or before that date; opening cash must never be added that way.
status: open
text_ref: docs/DESIGN.md#D-98
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-98
  - lane:strict
semantic_text: Opening lots added one at a time to an account must all be dated on its go-live date when it is seeded, on one date when it is not, and be refused for an account not seeded that has a current transaction on or before that date; opening cash must never be added that way.
logic_claims:
  - CLAIM-18B50858DCCF5E92
semantic_clauses:
  - Opening lots added one at a time to an account must all be dated on its go-live date when it is seeded, on one date when it is not, and be refused for an account not seeded that has a current transaction on or before that date; opening cash must never be added that way.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 7abb2d5a26a51db82bff5e1f790c2f9b43eeb2f5af7a4d73fcce775a15960ac2
semantic_inventory:
  - claim_key: CLAIM-18B50858DCCF5E92
    claim_text: Opening lots added one at a time to an account must all be dated on its go-live date when it is seeded, on one date when it is not, and be refused for an account not seeded that has a current transaction on or before that date; opening cash must never be added that way
    role: normative
    status: modeled
    span:
      start: 0
      end: 269
origin:
  kind: agent
  recorded_at: '2026-10-30T05:02:04.260Z'
id: REQ-ledger-opening-lots-one-at-a-time
type: req
---
Opening lots added one at a time to an account must all be dated on its go-live date when it is seeded, on one date when it is not, and be refused for an account not seeded that has a current transaction on or before that date; opening cash must never be added that way.

## Context

D-98 lets a manually entered account receive its opening lots one at a time. Keeping every lot on the go-live date so go-live never moves, and refusing a first lot that would count an earlier transaction twice, are this package's narrow reading. Recorded by the bonds implementation agent (impl/bonds-2).

## Source

> A manually entered account may receive its opening lots one at a time, and each is idempotent per series and purchase date.

docs/DESIGN.md#D-98

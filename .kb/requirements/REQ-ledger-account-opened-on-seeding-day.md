---
title: The owner must be able to mark an account as opened on its seeding day.
status: open
text_ref: design-decisions:D-168#1
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-168-1
  - lane:strict
semantic_text: The owner must be able to mark an account as opened on its seeding day.
logic_claims:
  - CLAIM-A8B3E65D68F2702E
semantic_clauses:
  - The owner must be able to mark an account as opened on its seeding day.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 7b34ade5b9383377cd6511c77c75eac6c246e2be4245f49ac44296b22c6f8e1f
semantic_inventory:
  - claim_key: CLAIM-A8B3E65D68F2702E
    claim_text: The owner must be able to mark an account as opened on its seeding day
    role: normative
    status: modeled
    span:
      start: 0
      end: 70
firmness: firm
required_evidence: asserted
origin:
  kind: import
  ref: design-decisions:D-168#1
  recorded_at: '2026-11-02T00:56:09.267Z'
id: REQ-ledger-account-opened-on-seeding-day
type: req
---
The owner must be able to mark an account as opened on its seeding day.

## Context

The author recorded it on 2026-11-01. The decision gives the purpose: an account so marked doesn't move the portfolio's default charts' start (D-138). No end-to-end test proves it yet, so its evidence floor is asserted.

## Source

> The owner can mark an account as opened on its seeding day, so its seeding doesn't move the portfolio's default start (D-138). An account seeded from a starting balance and not so marked moves it, since its earlier history is unknown.

OpenFief author decisions (docs/DESIGN.md, section 3.1), design-decisions:D-168#1

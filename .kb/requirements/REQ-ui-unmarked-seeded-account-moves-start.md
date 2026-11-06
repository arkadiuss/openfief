---
title: An account seeded from a starting balance and not marked as opened on its seeding day must move the portfolio's default charts' start.
status: open
text_ref: design-decisions:D-168#3
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-168-3
  - lane:strict
semantic_text: An account seeded from a starting balance and not marked as opened on its seeding day must move the portfolio's default charts' start.
logic_claims:
  - CLAIM-5A464C51EFEC4541
semantic_clauses:
  - An account seeded from a starting balance and not marked as opened on its seeding day must move the portfolio's default charts' start.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: be5be070ac186925b30dcad614dcda3db37a46139d8406bbd4df1011b26e36f8
semantic_inventory:
  - claim_key: CLAIM-5A464C51EFEC4541
    claim_text: An account seeded from a starting balance and not marked as opened on its seeding day must move the portfolio's default charts' start
    role: normative
    status: modeled
    span:
      start: 0
      end: 133
firmness: firm
required_evidence: asserted
origin:
  kind: import
  ref: design-decisions:D-168#3
  recorded_at: '2026-11-02T00:56:19.665Z'
id: REQ-ui-unmarked-seeded-account-moves-start
type: req
---
An account seeded from a starting balance and not marked as opened on its seeding day must move the portfolio's default charts' start.

## Context

The author recorded it on 2026-11-01. The decision gives the reason: its earlier history is unknown. No end-to-end test proves it yet, so its evidence floor is asserted.

## Source

> The owner can mark an account as opened on its seeding day, so its seeding doesn't move the portfolio's default start (D-138). An account seeded from a starting balance and not so marked moves it, since its earlier history is unknown.

OpenFief author decisions (docs/DESIGN.md, section 3.1), design-decisions:D-168#3

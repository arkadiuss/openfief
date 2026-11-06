---
title: An account marked as opened on its seeding day must not move the portfolio's default charts' start.
status: open
text_ref: design-decisions:D-168#2
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-168-2
  - lane:strict
semantic_text: An account marked as opened on its seeding day must not move the portfolio's default charts' start.
logic_claims:
  - CLAIM-42597B446FB62D8C
semantic_clauses:
  - An account marked as opened on its seeding day must not move the portfolio's default charts' start.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 34e769f50af0cb79706e1b7b76b2362517217187bef184a77429ac53e5c40057
semantic_inventory:
  - claim_key: CLAIM-42597B446FB62D8C
    claim_text: An account marked as opened on its seeding day must not move the portfolio's default charts' start
    role: normative
    status: modeled
    span:
      start: 0
      end: 98
firmness: firm
required_evidence: asserted
origin:
  kind: import
  ref: design-decisions:D-168#2
  recorded_at: '2026-11-02T00:56:14.414Z'
id: REQ-ui-opened-account-keeps-default-start
type: req
---
An account marked as opened on its seeding day must not move the portfolio's default charts' start.

## Context

The author recorded it on 2026-11-01. Reason not stated. D-138 starts the portfolio's charts by default on the latest day before which some account's history is unknown (REQ-ui-portfolio-chart-default-start); a marked account's history is whole. No end-to-end test proves it yet, so its evidence floor is asserted.

## Source

> The owner can mark an account as opened on its seeding day, so its seeding doesn't move the portfolio's default start (D-138). An account seeded from a starting balance and not so marked moves it, since its earlier history is unknown.

OpenFief author decisions (docs/DESIGN.md, section 3.1), design-decisions:D-168#2

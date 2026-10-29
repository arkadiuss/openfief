---
title: A manually entered account must be able to receive its opening lots one at a time, each idempotent per series and purchase date.
status: open
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-98-2
  - lane:strict
text_ref: design-decisions:D-98#2
origin:
  kind: import
  ref: design-decisions:D-98#2
  recorded_at: '2026-10-29T09:47:49.824Z'
semantic_text: A manually entered account must be able to receive its opening lots one at a time, each idempotent per series and purchase date.
logic_claims:
  - CLAIM-8A6EFA6532C51E49
semantic_clauses:
  - A manually entered account must be able to receive its opening lots one at a time, each idempotent per series and purchase date.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
firmness: firm
required_evidence: asserted
semantic_source_hash: 82cc7224b2fa7e33e2ff6291d0f293ee079a647d81e7ac1c3819b94eca73033e
semantic_inventory:
  - claim_key: CLAIM-8A6EFA6532C51E49
    claim_text: A manually entered account must be able to receive its opening lots one at a time, each idempotent per series and purchase date
    role: normative
    status: modeled
    span:
      start: 0
      end: 127
id: REQ-ledger-manual-opening-lots-one-at-a-time
type: req
---
A manually entered account must be able to receive its opening lots one at a time, each idempotent per series and purchase date.

## Context

A broker statement seeds an account with all its opening lots at once (D-23), but the owner enters treasury bond lots by hand one by one (D-9), and each lot bought before go-live is an opening lot (REQ-treasurybonds-pre-go-live-opening-lot). D-98 lets such a manually entered account receive its opening lots one at a time, each idempotent per series and purchase date, so entering the same lot again changes nothing. REQ-ledger-seed-refuse-reseed, which the author has not reviewed, refuses to seed an account that already holds an opening lot unless every incoming transaction is unchanged. The author recorded D-96 to D-99 after the design session that answered the implementation's questions.

## Source

> A treasury bond lot bought before go-live is an opening lot at go-live, keeping its own purchase date and cost (D-23). A manually entered account may receive its opening lots one at a time, and each is idempotent per series and purchase date. A lot bought after go-live is a deposit of its cost and a buy on its purchase date.

OpenFief author decisions (docs/DESIGN.md, section 3.2), design-decisions:D-98#2

---
title: Each sync request must carry each bound account's payout setting and the day it took effect, and a plugin must refuse an unknown setting or a paid-out one without a valid day.
status: closed
text_ref: docs/DESIGN.md#D-135
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-135
  - lane:strict
semantic_text: Each sync request must carry each bound account's payout setting and the day it took effect, and a plugin must refuse an unknown setting or a paid-out one without a valid day.
logic_claims:
  - CLAIM-E25C6ABB0EFC4888
semantic_clauses:
  - Each sync request must carry each bound account's payout setting and the day it took effect, and a plugin must refuse an unknown setting or a paid-out one without a valid day.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 5736572b161d902f727d2ac01bcec173558d3a010eed9fc60ac4b61cd2f8bef5
semantic_inventory:
  - claim_key: CLAIM-E25C6ABB0EFC4888
    claim_text: Each sync request must carry each bound account's payout setting and the day it took effect, and a plugin must refuse an unknown setting or a paid-out one without a valid day
    role: normative
    status: modeled
    span:
      start: 0
      end: 174
origin:
  kind: agent
  recorded_at: '2026-11-01T07:45:15.765Z'
id: REQ-ingest-sync-carries-payout
type: req
---
Each sync request must carry each bound account's payout setting and the day it took effect, and a plugin must refuse an unknown setting or a paid-out one without a valid day.

## Context

D-117's request carries what the connector needs to book what an account is paid; D-135 adds whether it is paid out. Recorded by the implementation agent of impl/bond-payouts; default firmness, not reviewed by the author yet.

## Source

> Each hand-entered bond account has a payout setting matching the option the bond registry offers: coupons and maturities stay as the account's cash, or are paid out. When they are paid out, each is booked as interest or a redemption and then a withdrawal of the same amount, so the account's cash matches the registry and the payout still counts as return. The treasury-bonds connector books maturities as it books coupons.

docs/DESIGN.md#D-135

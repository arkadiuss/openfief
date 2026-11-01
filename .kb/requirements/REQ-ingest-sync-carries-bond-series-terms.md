---
title: Each sync request must carry the stored terms of the bond series among its lots.
status: open
text_ref: design-decisions:D-143#2
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-143-2
  - lane:strict
semantic_text: Each sync request must carry the stored terms of the bond series among its lots.
logic_claims:
  - CLAIM-D91B1525AAF8EEF9
semantic_clauses:
  - Each sync request must carry the stored terms of the bond series among its lots.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: c0747ebadaec5588bca7ed21f586595be0e6d04b896b6d7a1ec91a61616ee905
semantic_inventory:
  - claim_key: CLAIM-D91B1525AAF8EEF9
    claim_text: Each sync request must carry the stored terms of the bond series among its lots
    role: normative
    status: modeled
    span:
      start: 0
      end: 79
firmness: firm
required_evidence: asserted
origin:
  kind: import
  ref: design-decisions:D-143#2
  recorded_at: '2026-11-01T12:25:30.849Z'
id: REQ-ingest-sync-carries-bond-series-terms
type: req
---
Each sync request must carry the stored terms of the bond series among its lots.

## Context

The author recorded it on 2026-11-01, answering the implementation's question #35. The owner enters a series' terms with a lot (D-97), and the treasury-bonds connector computes coupons from them with the Polish pack's bond rules (D-117), so the core holds no bond logic. No end-to-end test proves it yet, so its evidence floor is asserted.

## Source

> Besides the holdings and open lots, a sync request carries each bound account's tax wrapper and the stored terms of the bond series among its lots. The treasury-bonds connector books every coupon due since each lot's purchase on every sync, keyed per lot and coupon date, so a sync stays idempotent and a lot entered late gets its earlier coupons.

OpenFief author decisions (docs/DESIGN.md, section 3.3), design-decisions:D-143#2

---
title: Besides the holdings and open lots, each sync request must carry each bound account's tax wrapper.
status: open
text_ref: design-decisions:D-143#1
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-143-1
  - lane:strict
semantic_text: Besides the holdings and open lots, each sync request must carry each bound account's tax wrapper.
logic_claims:
  - CLAIM-6BA6447BC58A0A67
semantic_clauses:
  - Besides the holdings and open lots, each sync request must carry each bound account's tax wrapper.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 20a5e626d493100d9fda19700a02ce5c0938f1b955a1e90edd78c4a563a26665
semantic_inventory:
  - claim_key: CLAIM-6BA6447BC58A0A67
    claim_text: Besides the holdings and open lots, each sync request must carry each bound account's tax wrapper
    role: normative
    status: modeled
    span:
      start: 0
      end: 97
firmness: firm
required_evidence: asserted
origin:
  kind: import
  ref: design-decisions:D-143#1
  recorded_at: '2026-11-01T12:25:21.741Z'
id: REQ-ingest-sync-carries-tax-wrapper
type: req
---
Besides the holdings and open lots, each sync request must carry each bound account's tax wrapper.

## Context

The author recorded it on 2026-11-01, answering the implementation's question #35. D-117 has each sync request carry the holdings and open lots of the accounts bound to the instance (REQ-ingest-sync-carries-bound-holdings-lots); D-76 gives each account a tax wrapper. No end-to-end test proves it yet, so its evidence floor is asserted.

## Source

> Besides the holdings and open lots, a sync request carries each bound account's tax wrapper and the stored terms of the bond series among its lots. The treasury-bonds connector books every coupon due since each lot's purchase on every sync, keyed per lot and coupon date, so a sync stays idempotent and a lot entered late gets its earlier coupons.

OpenFief author decisions (docs/DESIGN.md, section 3.3), design-decisions:D-143#1

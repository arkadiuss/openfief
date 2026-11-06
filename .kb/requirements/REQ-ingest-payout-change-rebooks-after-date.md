---
title: After a change of an account's payout setting, the next sync must book the coupons and maturities after its date anew, superseding the earlier ones.
status: open
text_ref: design-decisions:D-162#3
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-162-3
  - lane:strict
semantic_text: After a change of an account's payout setting, the next sync must book the coupons and maturities after its date anew, superseding the earlier ones.
logic_claims:
  - CLAIM-B2544C5D2AF4F210
semantic_clauses:
  - After a change of an account's payout setting, the next sync must book the coupons and maturities after its date anew, superseding the earlier ones.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 4675d21ba32bc2150f14bcf33e7a3103d1209df299c52efd9303c1beb9c57621
semantic_inventory:
  - claim_key: CLAIM-B2544C5D2AF4F210
    claim_text: After a change of an account's payout setting, the next sync must book the coupons and maturities after its date anew, superseding the earlier ones
    role: normative
    status: modeled
    span:
      start: 0
      end: 147
firmness: firm
required_evidence: asserted
origin:
  kind: import
  ref: design-decisions:D-162#3
  recorded_at: '2026-11-02T00:55:29.589Z'
id: REQ-ingest-payout-change-rebooks-after-date
type: req
---
After a change of an account's payout setting, the next sync must book the coupons and maturities after its date anew, superseding the earlier ones.

## Why this changed

On 2026-11-01 the author recorded D-162. REQ-ingest-bond-payout-change-from-day and REQ-ingest-sync-carries-payout-history, default requirements the implementation agent recorded while building D-162 before any firm requirement for it existed, stated parts of this decision in their own words; the firm requirements for D-162 replace them, and they are closed. Their scenarios and the code that implements them move to the firm requirements, since the behavior built is unchanged.

## Context

The author recorded it on 2026-11-01. Reason not stated. Superseding is D-18's correction of the append-only log; the treasury-bonds connector books coupons and maturities on every sync (D-143). No end-to-end test proves it yet, so its evidence floor is asserted.

## Source

> An account's first payout setting covers its whole history. A change takes effect from a date the owner gives, today by default and possibly earlier, and the next sync books the coupons and maturities after that date anew, superseding the earlier ones (D-18, D-143).

OpenFief author decisions (docs/DESIGN.md, section 3.3), design-decisions:D-162#3

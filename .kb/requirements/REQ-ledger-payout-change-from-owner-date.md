---
title: A change of an account's payout setting must take effect from a date the owner gives, today by default and possibly earlier.
status: open
text_ref: design-decisions:D-162#2
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-162-2
  - lane:strict
semantic_text: A change of an account's payout setting must take effect from a date the owner gives, today by default and possibly earlier.
logic_claims:
  - CLAIM-3B7F5F175302A61D
semantic_clauses:
  - A change of an account's payout setting must take effect from a date the owner gives, today by default and possibly earlier.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: df7f03da4c71c778d34ad7bd07694f989d5a0e901d70d36aad0e0751dc508475
semantic_inventory:
  - claim_key: CLAIM-3B7F5F175302A61D
    claim_text: A change of an account's payout setting must take effect from a date the owner gives, today by default and possibly earlier
    role: normative
    status: modeled
    span:
      start: 0
      end: 123
firmness: firm
required_evidence: asserted
origin:
  kind: import
  ref: design-decisions:D-162#2
  recorded_at: '2026-11-02T00:55:21.162Z'
id: REQ-ledger-payout-change-from-owner-date
type: req
---
A change of an account's payout setting must take effect from a date the owner gives, today by default and possibly earlier.

## Why this changed

On 2026-11-01 the author recorded D-162. REQ-ledger-payout-setting-history, REQ-ingest-bond-payout-change-from-day and REQ-ui-bond-payout-change-date, default requirements the implementation agent recorded while building D-162 before any firm requirement for it existed, stated parts of this decision in their own words; the firm requirements for D-162 replace them, and they are closed. Their scenarios and the code that implements them move to the firm requirements, since the behavior built is unchanged.

## Context

The author recorded it on 2026-11-01. Reason not stated. The payout setting is D-135's (REQ-treasurybonds-account-payout-setting). No end-to-end test proves it yet, so its evidence floor is asserted.

## Source

> An account's first payout setting covers its whole history. A change takes effect from a date the owner gives, today by default and possibly earlier, and the next sync books the coupons and maturities after that date anew, superseding the earlier ones (D-18, D-143).

OpenFief author decisions (docs/DESIGN.md, section 3.3), design-decisions:D-162#2

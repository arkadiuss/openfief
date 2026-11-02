---
title: An account's first payout setting must cover its whole history.
status: open
text_ref: design-decisions:D-162#1
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-162-1
  - lane:strict
semantic_text: An account's first payout setting must cover its whole history.
logic_claims:
  - CLAIM-B7D70EDBB17DA0AF
semantic_clauses:
  - An account's first payout setting must cover its whole history.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 1cbfcf021cd85fd1c9f1a5c0d406eb84722b1d5049da4b80bdc2f7cb4d163a8d
semantic_inventory:
  - claim_key: CLAIM-B7D70EDBB17DA0AF
    claim_text: An account's first payout setting must cover its whole history
    role: normative
    status: modeled
    span:
      start: 0
      end: 62
firmness: firm
required_evidence: asserted
origin:
  kind: import
  ref: design-decisions:D-162#1
  recorded_at: '2026-11-02T00:55:13.988Z'
id: REQ-ledger-first-payout-setting-whole-history
type: req
---
An account's first payout setting must cover its whole history.

## Why this changed

On 2026-11-01 the author recorded D-162. REQ-ledger-payout-setting-history, a default requirement the implementation agent recorded while building D-162 before any firm requirement for it existed, stated this decision in its own words; the firm requirements for D-162 replace it, and it is closed. Its scenario and the code that implements it move to the firm requirements, since the behavior built is unchanged.

## Context

The author recorded it on 2026-11-01. Reason not stated. The payout setting is D-135's (REQ-treasurybonds-account-payout-setting). No end-to-end test proves it yet, so its evidence floor is asserted.

## Source

> An account's first payout setting covers its whole history. A change takes effect from a date the owner gives, today by default and possibly earlier, and the next sync books the coupons and maturities after that date anew, superseding the earlier ones (D-18, D-143).

OpenFief author decisions (docs/DESIGN.md, section 3.3), design-decisions:D-162#1

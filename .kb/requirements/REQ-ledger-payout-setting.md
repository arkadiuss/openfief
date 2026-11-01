---
title: Each account must keep a payout setting, cash by default or paid out, with the day it took effect, and what falls due must be paid out only when the setting is paid out and took effect on or before that day.
status: closed
text_ref: docs/DESIGN.md#D-135
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-135
  - lane:strict
semantic_text: Each account must keep a payout setting, cash by default or paid out, with the day it took effect, and what falls due must be paid out only when the setting is paid out and took effect on or before that day.
logic_claims:
  - CLAIM-C5BEE4A7399CD380
semantic_clauses:
  - Each account must keep a payout setting, cash by default or paid out, with the day it took effect, and what falls due must be paid out only when the setting is paid out and took effect on or before that day.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: abfc87eca7c984f6b94f2ecdbc9c84ab24bf5b8aecf707cd9c1944a0e1b4df1d
semantic_inventory:
  - claim_key: CLAIM-C5BEE4A7399CD380
    claim_text: Each account must keep a payout setting, cash by default or paid out, with the day it took effect, and what falls due must be paid out only when the setting is paid out and took effect on or before that day
    role: normative
    status: modeled
    span:
      start: 0
      end: 206
origin:
  kind: agent
  recorded_at: '2026-11-01T07:44:59.968Z'
id: REQ-ledger-payout-setting
type: req
---
Each account must keep a payout setting, cash by default or paid out, with the day it took effect, and what falls due must be paid out only when the setting is paid out and took effect on or before that day.

## Context

D-135 gives hand-entered bond accounts a payout setting. That a change takes effect on its day, so what was booked is never rewritten, is this implementation's reading (implementation question 55). Recorded by the implementation agent of impl/bond-payouts; default firmness, not reviewed by the author yet.

## Source

> Each hand-entered bond account has a payout setting matching the option the bond registry offers: coupons and maturities stay as the account's cash, or are paid out. When they are paid out, each is booked as interest or a redemption and then a withdrawal of the same amount, so the account's cash matches the registry and the payout still counts as return. The treasury-bonds connector books maturities as it books coupons.

docs/DESIGN.md#D-135

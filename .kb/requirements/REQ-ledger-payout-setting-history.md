---
title: An account's first payout setting must cover its whole history, a change must take effect from a day given, possibly earlier than today, and what falls due on a day must follow the setting in force on it.
status: open
text_ref: docs/DESIGN.md#D-162
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-162
  - lane:strict
semantic_text: An account's first payout setting must cover its whole history, a change must take effect from a day given, possibly earlier than today, and what falls due on a day must follow the setting in force on it.
logic_claims:
  - CLAIM-9D4B7C0F489D757B
semantic_clauses:
  - An account's first payout setting must cover its whole history, a change must take effect from a day given, possibly earlier than today, and what falls due on a day must follow the setting in force on it.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: e2f693b163e31b2f9452a6bf27dda3991ff452010f6a3e57e915645e317fc25e
semantic_inventory:
  - claim_key: CLAIM-9D4B7C0F489D757B
    claim_text: An account's first payout setting must cover its whole history, a change must take effect from a day given, possibly earlier than today, and what falls due on a day must follow the setting in force on it
    role: normative
    status: modeled
    span:
      start: 0
      end: 203
origin:
  kind: agent
  recorded_at: '2026-11-01T16:49:50.624Z'
id: REQ-ledger-payout-setting-history
type: req
---
An account's first payout setting must cover its whole history, a change must take effect from a day given, possibly earlier than today, and what falls due on a day must follow the setting in force on it.

## Context

D-162 has the first setting cover the whole history and a change take effect from a day the owner gives. Recorded by the implementation agent of D-159 to D-162 (impl/d159-162) with default firmness: no firm requirement for these decisions existed in kibi when they were built, so this traces the code to the decision directly until the design session imports the firm ones.

## Why this changed

The author decided this in D-162, which replaces the narrower reading REQ-ledger-payout-setting recorded before the decision existed, so that requirement is closed.

## Source

> An account's first payout setting covers its whole history. A change takes effect from a date the owner gives, today by default and possibly earlier, and the next sync books the coupons and maturities after that date anew, superseding the earlier ones (D-18, D-143).

docs/DESIGN.md#D-162

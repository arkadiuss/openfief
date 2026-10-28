---
title: A revaluation of an account from a date must rebuild its holding valuations for every day from that date, or its first transaction when later, to today, one per holding and day keyed by account, instrument, date and basis, replacing the stored ones from that date on, so that running it again gives the same rows.
status: open
text_ref: docs/DESIGN.md#D-36
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-36
  - lane:strict
semantic_text: A revaluation of an account from a date must rebuild its holding valuations for every day from that date, or its first transaction when later, to today, one per holding and day keyed by account, instrument, date and basis, replacing the stored ones from that date on, so that running it again gives the same rows.
logic_claims:
  - CLAIM-5EA93812331A6EF6
semantic_clauses:
  - A revaluation of an account from a date must rebuild its holding valuations for every day from that date, or its first transaction when later, to today, one per holding and day keyed by account, instrument, date and basis, replacing the stored ones from that date on, so that running it again gives the same rows.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 6e72bf51836682771c8747da2c9b6904660d8d7d4a344072689827d4ddd842d4
semantic_inventory:
  - claim_key: CLAIM-5EA93812331A6EF6
    claim_text: A revaluation of an account from a date must rebuild its holding valuations for every day from that date, or its first transaction when later, to today, one per holding and day keyed by account, instrument, date and basis, replacing the stored ones from that date on, so that running it again gives the same rows
    role: normative
    status: modeled
    span:
      start: 0
      end: 312
origin:
  kind: agent
  recorded_at: '2026-10-28T07:29:16.580Z'
id: REQ-valuation-revaluation-rows
type: req
---
A revaluation of an account from a date must rebuild its holding valuations for every day from that date, or its first transaction when later, to today, one per holding and day keyed by account, instrument, date and basis, replacing the stored ones from that date on, so that running it again gives the same rows.

## Context

D-36 makes valuations stored rows that can be rebuilt from the log and the inputs at any time; replacing every row from the start date makes a revaluation idempotent and lets a later change rebuild only what it affects. Refines REQ-decision-d36-1-stored-valuations and REQ-decision-d36-2-valuations-rebuildable. Modeled by Claude Code for the M0 revaluation task (D-36); the author has not reviewed it yet.

## Source

> Valuations are stored rows derived from the transaction log, prices, FX rates and index readings, and can be rebuilt from them at any time.

docs/DESIGN.md D-36

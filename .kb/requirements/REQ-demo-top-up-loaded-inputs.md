---
title: A demo load of another day without --reset-demo must first record again the valuation inputs of the dataset the demo marker names, leaving the ledger unchanged, and then fail asking for --reset-demo.
status: open
text_ref: docs/DESIGN.md#D-110
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-110
  - lane:strict
semantic_text: A demo load of another day without --reset-demo must first record again the valuation inputs of the dataset the demo marker names, leaving the ledger unchanged, and then fail asking for --reset-demo.
logic_claims:
  - CLAIM-9D8F1182326FCA6A
semantic_clauses:
  - A demo load of another day without --reset-demo must first record again the valuation inputs of the dataset the demo marker names, leaving the ledger unchanged, and then fail asking for --reset-demo.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 3a1f03807eaf24631e7b4689f73a75b590e9175efd9e3956a2622043841bf58b
semantic_inventory:
  - claim_key: CLAIM-9D8F1182326FCA6A
    claim_text: A demo load of another day without --reset-demo must first record again the valuation inputs of the dataset the demo marker names, leaving the ledger unchanged, and then fail asking for --reset-demo
    role: normative
    status: modeled
    span:
      start: 0
      end: 198
origin:
  kind: agent
  recorded_at: '2026-10-31T12:17:29.628Z'
id: REQ-demo-top-up-loaded-inputs
type: req
---
A demo load of another day without --reset-demo must first record again the valuation inputs of the dataset the demo marker names, leaving the ledger unchanged, and then fail asking for --reset-demo.

## Context

Refines REQ-demo-idempotent-per-day. The demo was loaded on 2026-10-28 by a loader that wrote no cross-check prices; every later run of the Job, with the loader that does, stopped at the refusal before recording any input, so the demo never showed the flagged price check D-110 asks the demo dataset to carry, nor the older FX rates. Every input is an upsert by its key, so recording the loaded dataset's inputs again changes nothing a load wrote and adds what a newer loader brings, and the running core's workers revalue. Modeled by Claude Code for the demo-numbers task: the author reported that the demo's IBKR holdings showed net and liquidation values as missing for want of an FX rate, and that the price checks page showed no check; the author has not reviewed this requirement yet.

## Source

> A cross-check price is flagged when it differs from the stored price by more than 1% of it, in either direction.

docs/DESIGN.md D-110

> The dataset also carries synthetic cross-check prices, one of them beyond the threshold, so the demo shows a flagged price check (D-110).

docs/dev/demo.md#demo-data

---
title: The core's attention checks must find flagged balance assertions, an instrument's latest flagged unacknowledged price check, batches with rows that can't be booked, unmapped broker accounts, unconfirmed tickers, a missing marginal rate an account's values need, and broken transfers.
status: open
text_ref: docs/DESIGN.md#D-130
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-130
  - lane:strict
semantic_text: The core's attention checks must find flagged balance assertions, an instrument's latest flagged unacknowledged price check, batches with rows that can't be booked, unmapped broker accounts, unconfirmed tickers, a missing marginal rate an account's values need, and broken transfers.
logic_claims:
  - CLAIM-A4849089CAADDBF3
semantic_clauses:
  - The core's attention checks must find flagged balance assertions, an instrument's latest flagged unacknowledged price check, batches with rows that can't be booked, unmapped broker accounts, unconfirmed tickers, a missing marginal rate an account's values need, and broken transfers.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: f865a0cf927b4a8f2720e60067f29e6b11d30f6d9dcd4c358655f6682fc8416f
semantic_inventory:
  - claim_key: CLAIM-A4849089CAADDBF3
    claim_text: The core's attention checks must find flagged balance assertions, an instrument's latest flagged unacknowledged price check, batches with rows that can't be booked, unmapped broker accounts, unconfirmed tickers, a missing marginal rate an account's values need, and broken transfers
    role: normative
    status: modeled
    span:
      start: 0
      end: 282
origin:
  kind: agent
  recorded_at: '2026-10-31T18:32:15.564Z'
id: REQ-attention-checks
type: req
---
The core's attention checks must find flagged balance assertions, an instrument's latest flagged unacknowledged price check, batches with rows that can't be booked, unmapped broker accounts, unconfirmed tickers, a missing marginal rate an account's values need, and broken transfers.

## Context

The attention items implementation (impl/attention) refined D-130's firm requirements into these rules; default firmness, the narrowest reading where D-130 is silent, with the open points in docs/implementation-questions.md (Attention items). Refines REQ-core-attention-item-checks for the checks that exist today.

## Source

> Every check that needs the owner raises an attention item in one append-only store: its kind, subject, state (open, acknowledged or resolved) and history. [...] Each change is also published as an event from the core's outbox, so notifications and, in later milestones, agents can act on them.

docs/DESIGN.md#D-130

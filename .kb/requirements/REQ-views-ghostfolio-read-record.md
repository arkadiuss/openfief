---
title: 'The Ghostfolio view must read the record only through the read API: every account, every page of its current transactions, and the stored valuations of its bonds from its first bond transaction through its valued-through date in ranges the read API accepts.'
status: open
text_ref: docs/DESIGN.md#D-6
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-6
  - lane:strict
semantic_text: 'The Ghostfolio view must read the record only through the read API: every account, every page of its current transactions, and the stored valuations of its bonds from its first bond transaction through its valued-through date in ranges the read API accepts.'
logic_claims:
  - CLAIM-B4F7B36F3267B462
semantic_clauses:
  - 'The Ghostfolio view must read the record only through the read API: every account, every page of its current transactions, and the stored valuations of its bonds from its first bond transaction through its valued-through date in ranges the read API accepts.'
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: c066a273ff7ad88724d282fb3f164095557e7cda165b0fcf7f6a4171b89a7895
semantic_inventory:
  - claim_key: CLAIM-B4F7B36F3267B462
    claim_text: 'The Ghostfolio view must read the record only through the read API: every account, every page of its current transactions, and the stored valuations of its bonds from its first bond transaction through its valued-through date in ranges the read API accepts'
    role: normative
    status: modeled
    span:
      start: 0
      end: 256
origin:
  kind: agent
  recorded_at: '2026-10-29T00:55:08.885Z'
id: REQ-views-ghostfolio-read-record
type: req
---
The Ghostfolio view must read the record only through the read API: every account, every page of its current transactions, and the stored valuations of its bonds from its first bond transaction through its valued-through date in ranges the read API accepts.

## Context

D-6: every view plugin uses the read API. Transactions come in pages of at most 500 and valuations in ranges of at most 3700 days, the read API's limits; only bonds' valuations are needed, for their price history (D-64). Modeled by Claude Code for the Ghostfolio view task; the author has not reviewed it yet.

## Source

> OpenFief's public API is read-only REST/JSON with an OpenAPI spec, versioned `/v1`, covering accounts, transactions, tax lots, holdings and valuations. Every view plugin, the built-in UI included, uses it.

docs/DESIGN.md D-6

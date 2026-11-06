---
title: 'The ledger must keep instruments'' display names per ISIN, append-only: the first name a connector delivers stays, the owner''s latest rename wins, names are cleaned and capped, and only the owner may rename, allowed to change every account that posts to the instrument.'
status: open
text_ref: docs/DESIGN.md#D-197
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-197
  - lane:strict
semantic_text: 'The ledger must keep instruments'' display names per ISIN, append-only: the first name a connector delivers stays, the owner''s latest rename wins, names are cleaned and capped, and only the owner may rename, allowed to change every account that posts to the instrument.'
logic_claims:
  - CLAIM-00543E24C6B0D741
semantic_clauses:
  - 'The ledger must keep instruments'' display names per ISIN, append-only: the first name a connector delivers stays, the owner''s latest rename wins, names are cleaned and capped, and only the owner may rename, allowed to change every account that posts to the instrument.'
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: cc875d276f9c93046e944e7b0f2cb1859fb213e7e64cf03875437779e6e41e0c
semantic_inventory:
  - claim_key: CLAIM-00543E24C6B0D741
    claim_text: 'The ledger must keep instruments'' display names per ISIN, append-only: the first name a connector delivers stays, the owner''s latest rename wins, names are cleaned and capped, and only the owner may rename, allowed to change every account that posts to the instrument'
    role: normative
    status: modeled
    span:
      start: 0
      end: 267
origin:
  kind: agent
  recorded_at: '2026-11-06T06:59:16.147Z'
id: REQ-ledger-instrument-names
type: req
---
The ledger must keep instruments' display names per ISIN, append-only: the first name a connector delivers stays, the owner's latest rename wins, names are cleaned and capped, and only the owner may rename, allowed to change every account that posts to the instrument.

## Context

Refines D-197; per ISIN rather than per listing, append-only like the rest of the ledger (D-18), and the authorizer asked as for every owner service (D-163). Modeled by Claude Code for the UI v2 overview slice; the author has not reviewed the wording.

## Source

> **[D-197]** An instrument has a display name: the first one a connector delivers for it, such as XTB's display name, IBKR's description or the market-data lookup's, which the owner can change. Pages show it before the ticker and ISIN.

docs/DESIGN.md#D-197

> **[D-163]** Authorization is per account and relationship-based. Every read path and owner service of the core asks one authorization interface whether the caller may act on each account.

docs/DESIGN.md#D-163

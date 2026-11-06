---
title: The TickerService must let only the owner list the confirmed tickers and correct one, asking first whether the owner may change every account the ticker's rows are in, and refuse a ticker never confirmed with FAILED_PRECONDITION and an unknown one with NOT_FOUND.
status: open
text_ref: docs/DESIGN.md#D-153
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-153
  - lane:strict
semantic_text: The TickerService must let only the owner list the confirmed tickers and correct one, asking first whether the owner may change every account the ticker's rows are in, and refuse a ticker never confirmed with FAILED_PRECONDITION and an unknown one with NOT_FOUND.
logic_claims:
  - CLAIM-38F3CE700270FE8D
semantic_clauses:
  - The TickerService must let only the owner list the confirmed tickers and correct one, asking first whether the owner may change every account the ticker's rows are in, and refuse a ticker never confirmed with FAILED_PRECONDITION and an unknown one with NOT_FOUND.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 1855ae9ee8c55c14f773f9a0fc4bc868f3e4d2a19065f9aff7fa54039e9575cc
semantic_inventory:
  - claim_key: CLAIM-38F3CE700270FE8D
    claim_text: The TickerService must let only the owner list the confirmed tickers and correct one, asking first whether the owner may change every account the ticker's rows are in, and refuse a ticker never confirmed with FAILED_PRECONDITION and an unknown one with NOT_FOUND
    role: normative
    status: modeled
    span:
      start: 0
      end: 262
origin:
  kind: agent
  recorded_at: '2026-11-02T07:00:52.008Z'
id: REQ-ingest-ticker-correction-service
type: req
---
The TickerService must let only the owner list the confirmed tickers and correct one, asking first whether the owner may change every account the ticker's rows are in, and refuse a ticker never confirmed with FAILED_PRECONDITION and an unknown one with NOT_FOUND.

## Context

D-153 lets the owner correct a confirmed ISIN; D-79 puts each kind of owner entry behind one owner-only service and D-163 has every owner service ask the authorizer about each account. The RPCs and their codes are this implementation's reading.

## Source

> **[D-153]** A ticker's confirmed ISIN (D-116) holds for every account at the same institution. The owner can correct a wrong one, which supersedes the transactions it booked (D-18).

docs/DESIGN.md#D-153

> **[D-163]** Authorization is per account and relationship-based. Every read path and owner service of the core asks one authorization interface whether the caller may act on each account.

docs/DESIGN.md#D-163

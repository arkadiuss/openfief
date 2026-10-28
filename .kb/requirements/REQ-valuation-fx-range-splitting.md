---
title: Fetching FX rates for a date range must ask the rate provider for consecutive ranges of at most its per-request day limit, split deterministically in date order, and return the combined rates sorted by date and currency.
status: open
text_ref: docs/DESIGN.md#D-20
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-20
  - lane:strict
semantic_text: Fetching FX rates for a date range must ask the rate provider for consecutive ranges of at most its per-request day limit, split deterministically in date order, and return the combined rates sorted by date and currency.
logic_claims:
  - CLAIM-084F2C5AC4B593DD
semantic_clauses:
  - Fetching FX rates for a date range must ask the rate provider for consecutive ranges of at most its per-request day limit, split deterministically in date order, and return the combined rates sorted by date and currency.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 279f3d003a53b6859913693fe00f11dd3f3a5b9bb7e185a6f7b4b363f8a17b90
semantic_inventory:
  - claim_key: CLAIM-084F2C5AC4B593DD
    claim_text: Fetching FX rates for a date range must ask the rate provider for consecutive ranges of at most its per-request day limit, split deterministically in date order, and return the combined rates sorted by date and currency
    role: normative
    status: modeled
    span:
      start: 0
      end: 219
origin:
  kind: agent
  recorded_at: '2026-10-28T05:54:24.104Z'
id: REQ-valuation-fx-range-splitting
type: req
---
Fetching FX rates for a date range must ask the rate provider for consecutive ranges of at most its per-request day limit, split deterministically in date order, and return the combined rates sorted by date and currency.

## Context

NBP's Web API serves at most 93 days per request, so loading a longer history (for example a year of rates for tax) needs several requests. Splitting is part of the use case, not the adapter, so a swappable provider (D-20) only declares its limit, and the split depends only on the range and the limit so that replayed requests stay stable. Refines REQ-AUTO-D1568676DF81CAF7. Modeled by Claude Code for the M0 NBP FX source task; the author has not reviewed it yet.

## Source

> In M0, NBP table A rates convert amounts to PLN, behind a swappable rate-provider interface with NBP as the default FX plugin.

docs/DESIGN.md D-20

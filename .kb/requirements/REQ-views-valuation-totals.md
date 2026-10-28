---
title: Valuation totals must be one per date, summing the stored valuations dated on it exactly in the reporting currency and separately per original currency, and must refuse a date range that is reversed or longer than 3700 days.
status: open
text_ref: docs/DESIGN.md#D-30
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-30
  - lane:strict
semantic_text: Valuation totals must be one per date, summing the stored valuations dated on it exactly in the reporting currency and separately per original currency, and must refuse a date range that is reversed or longer than 3700 days.
logic_claims:
  - CLAIM-E38586693C1E5BC8
semantic_clauses:
  - Valuation totals must be one per date, summing the stored valuations dated on it exactly in the reporting currency and separately per original currency, and must refuse a date range that is reversed or longer than 3700 days.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: a8981d623b66f431e82e71171622c52b76911810161d2fd29b62cb248544c279
semantic_inventory:
  - claim_key: CLAIM-E38586693C1E5BC8
    claim_text: Valuation totals must be one per date, summing the stored valuations dated on it exactly in the reporting currency and separately per original currency, and must refuse a date range that is reversed or longer than 3700 days
    role: normative
    status: modeled
    span:
      start: 0
      end: 223
origin:
  kind: agent
  recorded_at: '2026-10-28T07:18:28.813Z'
id: REQ-views-valuation-totals
type: req
---
Valuation totals must be one per date, summing the stored valuations dated on it exactly in the reporting currency and separately per original currency, and must refuse a date range that is reversed or longer than 3700 days.

## Context

Feeds the value-over-time chart (firm REQ-AUTO-718AA30AAFD0E6E8) and keeps amounts in their original currency next to the PLN value (firm REQ-AUTO-5CB834BAA25E2C4C, D-20). The range limit, about ten years of daily values, bounds one request's work; it is the agent's choice, not the author's. Modeled by Claude Code for the M0 read API task; the author has not reviewed it yet.

## Source

> The built-in UI draws two charts in M0: allocation (by asset class, currency, broker or sleeve) and value over time, which starts at go-live (D-23).

docs/DESIGN.md D-30

> Every amount keeps its original currency. In M0, NBP table A rates convert amounts to PLN, behind a swappable rate-provider interface with NBP as the default FX plugin.

docs/DESIGN.md D-20

---
title: The allocation must count each holding at its stored valuation on the date, list a holding without one as unvalued, and convert each cash balance to PLN at the latest FX rate dated on or before the date, rounded half-even to grosze, failing when a cash currency has no rate.
status: open
text_ref: docs/DESIGN.md#D-30
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-30
  - lane:strict
semantic_text: The allocation must count each holding at its stored valuation on the date, list a holding without one as unvalued, and convert each cash balance to PLN at the latest FX rate dated on or before the date, rounded half-even to grosze, failing when a cash currency has no rate.
logic_claims:
  - CLAIM-47F6FBCADD21B983
semantic_clauses:
  - The allocation must count each holding at its stored valuation on the date, list a holding without one as unvalued, and convert each cash balance to PLN at the latest FX rate dated on or before the date, rounded half-even to grosze, failing when a cash currency has no rate.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: e26ecbdd9141521c8a39c0ef950852418ea37a59206467d9a36d694003937d2e
semantic_inventory:
  - claim_key: CLAIM-47F6FBCADD21B983
    claim_text: The allocation must count each holding at its stored valuation on the date, list a holding without one as unvalued, and convert each cash balance to PLN at the latest FX rate dated on or before the date, rounded half-even to grosze, failing when a cash currency has no rate
    role: normative
    status: modeled
    span:
      start: 0
      end: 273
origin:
  kind: agent
  recorded_at: '2026-10-28T07:26:09.267Z'
id: REQ-views-allocation-inputs
type: req
---
The allocation must count each holding at its stored valuation on the date, list a holding without one as unvalued, and convert each cash balance to PLN at the latest FX rate dated on or before the date, rounded half-even to grosze, failing when a cash currency has no rate.

## Context

Feeds the allocation chart (firm REQ-AUTO-68636F42D98D4CF7, D-30). Holdings use the stored valuations (D-36), so the chart agrees with the valuation rows; cash has no stored valuation, so it is converted here with the rates D-20 names, through a consumer-side FXRates port. Half-even rounding of cash is the agent's choice; no decision names a mode for display conversion. Failing on a missing rate keeps the total from silently shrinking. Modeled by Claude Code for the M0 read API task; the author has not reviewed it yet.

## Source

> The built-in UI draws two charts in M0: allocation (by asset class, currency, broker or sleeve) and value over time, which starts at go-live (D-23).

docs/DESIGN.md D-30

> Every amount keeps its original currency. In M0, NBP table A rates convert amounts to PLN, behind a swappable rate-provider interface with NBP as the default FX plugin.

docs/DESIGN.md D-20

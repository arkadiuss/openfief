---
title: The allocation must count each holding at its stored valuation on the date, list a holding without one as unvalued, and convert each cash balance to PLN at the latest FX rate dated on or before the date, rounded half up to grosze as holding values are, failing when a cash currency has no rate.
status: open
text_ref: docs/DESIGN.md#D-30
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-30
  - lane:strict
semantic_text: The allocation must count each holding at its stored valuation on the date, list a holding without one as unvalued, and convert each cash balance to PLN at the latest FX rate dated on or before the date, rounded half up to grosze as holding values are, failing when a cash currency has no rate.
logic_claims:
  - CLAIM-2F1750F1D6817487
semantic_clauses:
  - The allocation must count each holding at its stored valuation on the date, list a holding without one as unvalued, and convert each cash balance to PLN at the latest FX rate dated on or before the date, rounded half up to grosze as holding values are, failing when a cash currency has no rate.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 288233da0a3aeeb0d7998e8d1d1336c42b5b5b8310b6f43409515144dfbdc9db
semantic_inventory:
  - claim_key: CLAIM-2F1750F1D6817487
    claim_text: The allocation must count each holding at its stored valuation on the date, list a holding without one as unvalued, and convert each cash balance to PLN at the latest FX rate dated on or before the date, rounded half up to grosze as holding values are, failing when a cash currency has no rate
    role: normative
    status: modeled
    span:
      start: 0
      end: 293
origin:
  kind: agent
  recorded_at: '2026-10-28T07:26:09.267Z'
id: REQ-views-allocation-inputs
type: req
---
The allocation must count each holding at its stored valuation on the date, list a holding without one as unvalued, and convert each cash balance to PLN at the latest FX rate dated on or before the date, rounded half up to grosze as holding values are, failing when a cash currency has no rate.

## Context

D-30 asks for the allocation chart; D-62 puts totals in the reporting currency. Cash has no stored valuation, so the read API converts it at the FX rate for the date. It rounds with valuation.ValueRounding, half up, the rounding the valuation context uses for holding values, so cash and holdings add up alike; no decision fixes the rounding mode. Corrected from an earlier text that said half-even, which the code never did.

## Source

> **[D-30]** The built-in UI draws two charts in M0: allocation (by asset class, currency, broker or sleeve) and value over time, which starts at go-live (D-23).

docs/DESIGN.md#D-30

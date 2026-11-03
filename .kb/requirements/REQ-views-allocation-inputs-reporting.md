---
title: The allocation must count each holding at its stored valuation on the date, list a holding without one as unvalued, and convert each cash balance to the reporting currency at the latest FX rate into it dated on or before the date, rounded half up to the minor unit as holding values are, failing when a cash currency has no rate.
status: open
text_ref: docs/DESIGN.md#D-180
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-180
  - lane:strict
semantic_text: The allocation must count each holding at its stored valuation on the date, list a holding without one as unvalued, and convert each cash balance to the reporting currency at the latest FX rate into it dated on or before the date, rounded half up to the minor unit as holding values are, failing when a cash currency has no rate.
logic_claims:
  - CLAIM-41E7208C7B0F0EF3
semantic_clauses:
  - The allocation must count each holding at its stored valuation on the date, list a holding without one as unvalued, and convert each cash balance to the reporting currency at the latest FX rate into it dated on or before the date, rounded half up to the minor unit as holding values are, failing when a cash currency has no rate.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: f2781a74c84bc249276c9c5e8d5bf4e1f38feadfcf0d9ccaced7e18df83e7a8f
semantic_inventory:
  - claim_key: CLAIM-41E7208C7B0F0EF3
    claim_text: The allocation must count each holding at its stored valuation on the date, list a holding without one as unvalued, and convert each cash balance to the reporting currency at the latest FX rate into it dated on or before the date, rounded half up to the minor unit as holding values are, failing when a cash currency has no rate
    role: normative
    status: modeled
    span:
      start: 0
      end: 328
origin:
  kind: agent
  recorded_at: '2026-11-02T23:55:59.322Z'
id: REQ-views-allocation-inputs-reporting
type: req
---
The allocation must count each holding at its stored valuation on the date, list a holding without one as unvalued, and convert each cash balance to the reporting currency at the latest FX rate into it dated on or before the date, rounded half up to the minor unit as holding values are, failing when a cash currency has no rate.

## Why this changed

On 2026-11-02 the author recorded D-180: stored valuations name their currency and stored exchange rates their base currency, both taken from the jurisdiction pack, and neither is assumed to be PLN. REQ-views-allocation-inputs stated it in PLN ("The allocation must count each holding at its stored valuation on the date, list a holding without one as unvalued, and convert each cash balance to PLN at the latest FX rate dated on or before the date, rounded half up to grosze as holding values are, failing when a cash currency has no rate."); this requirement replaces it, naming the currency instead of assuming PLN.

## Context

The Polish pack's reporting currency is PLN, so its behavior is unchanged; another pack's currency flows the same way, and a reporting currency the pack can't convert to stays refused (T-12).

## Source

> Stored valuations name their currency and stored exchange rates their base currency, both taken from the jurisdiction pack; neither is assumed to be PLN. A reporting currency the pack can't convert to stays refused until T-12 is resolved.

OpenFief author decisions (docs/DESIGN.md, section 3.3), design-decisions:D-180#1


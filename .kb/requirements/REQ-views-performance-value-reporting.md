---
title: Each day's portfolio value in the performance series must be its holdings at their stored valuations plus every account's cash converted to the reporting currency at the latest FX rate into it on or before the day and rounded half up to the minor unit, plus money in transit between its accounts, with each opening lot flowing in at its share of its holding's valuation.
status: open
text_ref: docs/DESIGN.md#D-180
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-180
  - lane:strict
semantic_text: Each day's portfolio value in the performance series must be its holdings at their stored valuations plus every account's cash converted to the reporting currency at the latest FX rate into it on or before the day and rounded half up to the minor unit, plus money in transit between its accounts, with each opening lot flowing in at its share of its holding's valuation.
logic_claims:
  - CLAIM-DC90EE1809242CBB
semantic_clauses:
  - Each day's portfolio value in the performance series must be its holdings at their stored valuations plus every account's cash converted to the reporting currency at the latest FX rate into it on or before the day and rounded half up to the minor unit, plus money in transit between its accounts, with each opening lot flowing in at its share of its holding's valuation.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 08badbe9c82d31022465184893f619cddb61bcbc8fa9886b667170b132926eee
semantic_inventory:
  - claim_key: CLAIM-DC90EE1809242CBB
    claim_text: Each day's portfolio value in the performance series must be its holdings at their stored valuations plus every account's cash converted to the reporting currency at the latest FX rate into it on or before the day and rounded half up to the minor unit, plus money in transit between its accounts, with each opening lot flowing in at its share of its holding's valuation
    role: normative
    status: modeled
    span:
      start: 0
      end: 369
origin:
  kind: agent
  recorded_at: '2026-11-02T23:56:10.708Z'
id: REQ-views-performance-value-reporting
type: req
---
Each day's portfolio value in the performance series must be its holdings at their stored valuations plus every account's cash converted to the reporting currency at the latest FX rate into it on or before the day and rounded half up to the minor unit, plus money in transit between its accounts, with each opening lot flowing in at its share of its holding's valuation.

## Why this changed

On 2026-11-02 the author recorded D-180: stored valuations name their currency and stored exchange rates their base currency, both taken from the jurisdiction pack, and neither is assumed to be PLN. REQ-views-performance-value stated it in PLN ("Each day's portfolio value in the performance series must be its holdings at their stored valuations plus every account's cash converted to PLN at the latest FX rate on or before the day and rounded half up to grosze, plus money in transit between its accounts, with each opening lot flowing in at its share of its holding's valuation."); this requirement replaces it, naming the currency instead of assuming PLN.

## Context

The Polish pack's reporting currency is PLN, so its behavior is unchanged; another pack's currency flows the same way, and a reporting currency the pack can't convert to stays refused (T-12).

## Source

> Stored valuations name their currency and stored exchange rates their base currency, both taken from the jurisdiction pack; neither is assumed to be PLN. A reporting currency the pack can't convert to stays refused until T-12 is resolved.

OpenFief author decisions (docs/DESIGN.md, section 3.3), design-decisions:D-180#1


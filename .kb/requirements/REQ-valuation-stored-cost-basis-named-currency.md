---
title: A stored valuation of a holding priced at market must keep, with its net value, its cost basis for tax and its sale cost in the currency the valuation names.
status: open
text_ref: docs/DESIGN.md#D-180
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-180
  - lane:strict
semantic_text: A stored valuation of a holding priced at market must keep, with its net value, its cost basis for tax and its sale cost in the currency the valuation names.
logic_claims:
  - CLAIM-724853CCFFA63191
semantic_clauses:
  - A stored valuation of a holding priced at market must keep, with its net value, its cost basis for tax and its sale cost in the currency the valuation names.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 4ff471f4b93cd71409cd842cc8fd06bd419aa546fc5ab2446067b9870ea2e508
semantic_inventory:
  - claim_key: CLAIM-724853CCFFA63191
    claim_text: A stored valuation of a holding priced at market must keep, with its net value, its cost basis for tax and its sale cost in the currency the valuation names
    role: normative
    status: modeled
    span:
      start: 0
      end: 156
origin:
  kind: agent
  recorded_at: '2026-11-02T23:55:48.275Z'
id: REQ-valuation-stored-cost-basis-named-currency
type: req
---
A stored valuation of a holding priced at market must keep, with its net value, its cost basis for tax and its sale cost in the currency the valuation names.

## Why this changed

On 2026-11-02 the author recorded D-180: stored valuations name their currency and stored exchange rates their base currency, both taken from the jurisdiction pack, and neither is assumed to be PLN. REQ-valuation-stored-cost-basis stated it in PLN ("A stored valuation of a holding priced at market must keep, with its net value, its cost basis for tax and its sale cost in PLN."); this requirement replaces it, naming the currency instead of assuming PLN.

## Context

The Polish pack's reporting currency is PLN, so its behavior is unchanged; another pack's currency flows the same way, and a reporting currency the pack can't convert to stays refused (T-12).

## Source

> Stored valuations name their currency and stored exchange rates their base currency, both taken from the jurisdiction pack; neither is assumed to be PLN. A reporting currency the pack can't convert to stays refused until T-12 is resolved.

OpenFief author decisions (docs/DESIGN.md, section 3.3), design-decisions:D-180#1


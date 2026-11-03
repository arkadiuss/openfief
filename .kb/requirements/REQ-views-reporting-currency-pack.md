---
title: The read API must report totals, valuations and the allocation in the configured reporting currency next to each value's original currency, and must refuse to start with a reporting currency other than the one the jurisdiction pack's stored valuations and rates are in, with a typed unsupported reporting currency error, until cross rates are decided.
status: open
text_ref: docs/DESIGN.md#D-180
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-180
  - lane:strict
semantic_text: The read API must report totals, valuations and the allocation in the configured reporting currency next to each value's original currency, and must refuse to start with a reporting currency other than the one the jurisdiction pack's stored valuations and rates are in, with a typed unsupported reporting currency error, until cross rates are decided.
logic_claims:
  - CLAIM-F66E951806B1CE8C
semantic_clauses:
  - The read API must report totals, valuations and the allocation in the configured reporting currency next to each value's original currency, and must refuse to start with a reporting currency other than the one the jurisdiction pack's stored valuations and rates are in, with a typed unsupported reporting currency error, until cross rates are decided.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 5646d133dd6c5564b4ab7770cb0347c319a4de641175a09a28f3535f64353f99
semantic_inventory:
  - claim_key: CLAIM-F66E951806B1CE8C
    claim_text: The read API must report totals, valuations and the allocation in the configured reporting currency next to each value's original currency, and must refuse to start with a reporting currency other than the one the jurisdiction pack's stored valuations and rates are in, with a typed unsupported reporting currency error, until cross rates are decided
    role: normative
    status: modeled
    span:
      start: 0
      end: 350
origin:
  kind: agent
  recorded_at: '2026-11-02T23:55:07.367Z'
id: REQ-views-reporting-currency-pack
type: req
---
The read API must report totals, valuations and the allocation in the configured reporting currency next to each value's original currency, and must refuse to start with a reporting currency other than the one the jurisdiction pack's stored valuations and rates are in, with a typed unsupported reporting currency error, until cross rates are decided.

## Why this changed

On 2026-11-02 the author recorded D-180: stored valuations name their currency and stored exchange rates their base currency, both taken from the jurisdiction pack, and neither is assumed to be PLN. REQ-views-reporting-currency stated it in PLN ("The read API must report totals, valuations and the allocation in the configured reporting currency next to each value's original currency, and must refuse to start with a reporting currency other than PLN with a typed unsupported reporting currency error until cross rates are decided."); this requirement replaces it, naming the currency instead of assuming PLN.

## Context

The Polish pack's reporting currency is PLN, so its behavior is unchanged; another pack's currency flows the same way, and a reporting currency the pack can't convert to stays refused (T-12).

## Source

> Stored valuations name their currency and stored exchange rates their base currency, both taken from the jurisdiction pack; neither is assumed to be PLN. A reporting currency the pack can't convert to stays refused until T-12 is resolved.

OpenFief author decisions (docs/DESIGN.md, section 3.3), design-decisions:D-180#1


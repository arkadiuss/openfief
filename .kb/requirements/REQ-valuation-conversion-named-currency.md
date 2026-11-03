---
title: Converting an amount into a currency must multiply it exactly by the rate of its own currency into that currency and round the result once to the minor unit with a rounding mode the caller states, leaving the original amount unchanged.
status: open
text_ref: docs/DESIGN.md#D-180
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-180
  - lane:strict
semantic_text: Converting an amount into a currency must multiply it exactly by the rate of its own currency into that currency and round the result once to the minor unit with a rounding mode the caller states, leaving the original amount unchanged.
logic_claims:
  - CLAIM-EF5DC6FD028EE414
semantic_clauses:
  - Converting an amount into a currency must multiply it exactly by the rate of its own currency into that currency and round the result once to the minor unit with a rounding mode the caller states, leaving the original amount unchanged.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: e1cfcf5e4fb80d559febd77b34c6b064c9fea633201e106252edd85bb1b7ac64
semantic_inventory:
  - claim_key: CLAIM-EF5DC6FD028EE414
    claim_text: Converting an amount into a currency must multiply it exactly by the rate of its own currency into that currency and round the result once to the minor unit with a rounding mode the caller states, leaving the original amount unchanged
    role: normative
    status: modeled
    span:
      start: 0
      end: 234
origin:
  kind: agent
  recorded_at: '2026-11-02T23:54:58.068Z'
id: REQ-valuation-conversion-named-currency
type: req
---
Converting an amount into a currency must multiply it exactly by the rate of its own currency into that currency and round the result once to the minor unit with a rounding mode the caller states, leaving the original amount unchanged.

## Why this changed

On 2026-11-02 the author recorded D-180: stored valuations name their currency and stored exchange rates their base currency, both taken from the jurisdiction pack, and neither is assumed to be PLN. REQ-valuation-pln-conversion stated it in PLN ("Converting an amount to PLN must multiply it exactly by the rate for its own currency and round the result once to grosze with a rounding mode the caller states, leaving the original amount unchanged."); this requirement replaces it, naming the currency instead of assuming PLN.

## Context

The Polish pack's reporting currency is PLN, so its behavior is unchanged; another pack's currency flows the same way, and a reporting currency the pack can't convert to stays refused (T-12).

## Source

> Stored valuations name their currency and stored exchange rates their base currency, both taken from the jurisdiction pack; neither is assumed to be PLN. A reporting currency the pack can't convert to stays refused until T-12 is resolved.

OpenFief author decisions (docs/DESIGN.md, section 3.3), design-decisions:D-180#1


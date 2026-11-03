---
title: Recording a price that is not positive or names no instrument, an FX rate that is not positive, names no base or is quoted in its own currency, an index reading of an unknown kind or month, bond terms without an ISIN or series, or an unknown tax wrapper must fail and store nothing.
status: open
text_ref: docs/DESIGN.md#D-180
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-180
  - lane:strict
semantic_text: Recording a price that is not positive or names no instrument, an FX rate that is not positive, names no base or is quoted in its own currency, an index reading of an unknown kind or month, bond terms without an ISIN or series, or an unknown tax wrapper must fail and store nothing.
logic_claims:
  - CLAIM-73E2AB4610F596C5
semantic_clauses:
  - Recording a price that is not positive or names no instrument, an FX rate that is not positive, names no base or is quoted in its own currency, an index reading of an unknown kind or month, bond terms without an ISIN or series, or an unknown tax wrapper must fail and store nothing.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: a61ccfd3860e07711f3fbe15ce6fc611e4f6c8004cb34a9bb543a22789ee6e11
semantic_inventory:
  - claim_key: CLAIM-73E2AB4610F596C5
    claim_text: Recording a price that is not positive or names no instrument, an FX rate that is not positive, names no base or is quoted in its own currency, an index reading of an unknown kind or month, bond terms without an ISIN or series, or an unknown tax wrapper must fail and store nothing
    role: normative
    status: modeled
    span:
      start: 0
      end: 281
origin:
  kind: agent
  recorded_at: '2026-11-02T23:55:18.569Z'
id: REQ-valuation-input-validation-named-base
type: req
---
Recording a price that is not positive or names no instrument, an FX rate that is not positive, names no base or is quoted in its own currency, an index reading of an unknown kind or month, bond terms without an ISIN or series, or an unknown tax wrapper must fail and store nothing.

## Why this changed

On 2026-11-02 the author recorded D-180: stored valuations name their currency and stored exchange rates their base currency, both taken from the jurisdiction pack, and neither is assumed to be PLN. REQ-valuation-input-validation stated it in PLN ("Recording a price that is not positive or names no instrument, an FX rate that is not positive or is for PLN, an index reading of an unknown kind or month, bond terms without an ISIN or series, or an unknown tax wrapper must fail and store nothing."); this requirement replaces it, naming the currency instead of assuming PLN.

## Context

The Polish pack's reporting currency is PLN, so its behavior is unchanged; another pack's currency flows the same way, and a reporting currency the pack can't convert to stays refused (T-12).

## Source

> Stored valuations name their currency and stored exchange rates their base currency, both taken from the jurisdiction pack; neither is assumed to be PLN. A reporting currency the pack can't convert to stays refused until T-12 is resolved.

OpenFief author decisions (docs/DESIGN.md, section 3.3), design-decisions:D-180#1


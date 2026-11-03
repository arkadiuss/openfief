---
title: A holding of an instrument without bond series terms must be valued at the latest end-of-day price of the instrument on or before the day times its quantity, in the price's currency, and converted to the jurisdiction pack's reporting currency at the latest FX rate into it on or before the day, rounded half up to the minor unit.
status: open
text_ref: docs/DESIGN.md#D-180
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-180
  - lane:strict
semantic_text: A holding of an instrument without bond series terms must be valued at the latest end-of-day price of the instrument on or before the day times its quantity, in the price's currency, and converted to the jurisdiction pack's reporting currency at the latest FX rate into it on or before the day, rounded half up to the minor unit.
logic_claims:
  - CLAIM-2BBB82D57050EC4D
semantic_clauses:
  - A holding of an instrument without bond series terms must be valued at the latest end-of-day price of the instrument on or before the day times its quantity, in the price's currency, and converted to the jurisdiction pack's reporting currency at the latest FX rate into it on or before the day, rounded half up to the minor unit.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: c7b108e1801d2101e27fe170e98f3e1cef3f5965050e70f1bc6caf6ca26d2e3c
semantic_inventory:
  - claim_key: CLAIM-2BBB82D57050EC4D
    claim_text: A holding of an instrument without bond series terms must be valued at the latest end-of-day price of the instrument on or before the day times its quantity, in the price's currency, and converted to the jurisdiction pack's reporting currency at the latest FX rate into it on or before the day, rounded half up to the minor unit
    role: normative
    status: modeled
    span:
      start: 0
      end: 328
origin:
  kind: agent
  recorded_at: '2026-11-02T23:55:27.773Z'
id: REQ-valuation-market-holding-pack-currency
type: req
---
A holding of an instrument without bond series terms must be valued at the latest end-of-day price of the instrument on or before the day times its quantity, in the price's currency, and converted to the jurisdiction pack's reporting currency at the latest FX rate into it on or before the day, rounded half up to the minor unit.

## Why this changed

On 2026-11-02 the author recorded D-180: stored valuations name their currency and stored exchange rates their base currency, both taken from the jurisdiction pack, and neither is assumed to be PLN. REQ-valuation-market-holding stated it in PLN ("A holding of an instrument without bond series terms must be valued at the latest end-of-day price of the instrument on or before the day times its quantity, in the price's currency, and converted to PLN at the latest FX rate on or before the day, rounded half up to grosze."); this requirement replaces it, naming the currency instead of assuming PLN.

## Context

The Polish pack's reporting currency is PLN, so its behavior is unchanged; another pack's currency flows the same way, and a reporting currency the pack can't convert to stays refused (T-12).

## Source

> Stored valuations name their currency and stored exchange rates their base currency, both taken from the jurisdiction pack; neither is assumed to be PLN. A reporting currency the pack can't convert to stays refused until T-12 is resolved.

OpenFief author decisions (docs/DESIGN.md, section 3.3), design-decisions:D-180#1


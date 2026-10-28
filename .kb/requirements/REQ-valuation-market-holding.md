---
title: A holding of an instrument without bond series terms must be valued at the latest end-of-day price of the instrument on or before the day times its quantity, in the price's currency, and converted to PLN at the latest FX rate on or before the day, rounded half up to grosze.
status: open
text_ref: docs/DESIGN.md#D-21
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-21
  - lane:strict
semantic_text: A holding of an instrument without bond series terms must be valued at the latest end-of-day price of the instrument on or before the day times its quantity, in the price's currency, and converted to PLN at the latest FX rate on or before the day, rounded half up to grosze.
logic_claims:
  - CLAIM-396F581CD11DACD3
semantic_clauses:
  - A holding of an instrument without bond series terms must be valued at the latest end-of-day price of the instrument on or before the day times its quantity, in the price's currency, and converted to PLN at the latest FX rate on or before the day, rounded half up to grosze.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: ffff05498aff2d3feba0fee5d83531073e6b542cbbe15510280b3adcee3bdb07
semantic_inventory:
  - claim_key: CLAIM-396F581CD11DACD3
    claim_text: A holding of an instrument without bond series terms must be valued at the latest end-of-day price of the instrument on or before the day times its quantity, in the price's currency, and converted to PLN at the latest FX rate on or before the day, rounded half up to grosze
    role: normative
    status: modeled
    span:
      start: 0
      end: 273
origin:
  kind: agent
  recorded_at: '2026-10-28T07:29:47.489Z'
id: REQ-valuation-market-holding
type: req
---
A holding of an instrument without bond series terms must be valued at the latest end-of-day price of the instrument on or before the day times its quantity, in the price's currency, and converted to PLN at the latest FX rate on or before the day, rounded half up to grosze.

## Context

Prices are end of day (D-21) and amounts keep their currency with NBP rates converting to PLN (D-20); taking the latest price and rate on or before the day values weekends and holidays. The half-up rounding of the PLN value is not decided by the author; it matches the bond tables (D-37). Modeled by Claude Code for the M0 revaluation task (D-36); the author has not reviewed it yet.

## Source

> Prices are end of day in M0.

docs/DESIGN.md D-21

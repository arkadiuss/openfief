---
title: The FX rate used for a currency on a date must be the latest rate for that currency dated on or before the date, and a missing rate or two different rates for the chosen date must be an error.
status: open
text_ref: docs/DESIGN.md#D-20
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-20
  - lane:strict
semantic_text: The FX rate used for a currency on a date must be the latest rate for that currency dated on or before the date, and a missing rate or two different rates for the chosen date must be an error.
logic_claims:
  - CLAIM-52A0171A2C48C99D
semantic_clauses:
  - The FX rate used for a currency on a date must be the latest rate for that currency dated on or before the date, and a missing rate or two different rates for the chosen date must be an error.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: cfde262bc19472c4f8bb7f58526c20a8edda3308405e3a9932bb3608c3795a88
semantic_inventory:
  - claim_key: CLAIM-52A0171A2C48C99D
    claim_text: The FX rate used for a currency on a date must be the latest rate for that currency dated on or before the date, and a missing rate or two different rates for the chosen date must be an error
    role: normative
    status: modeled
    span:
      start: 0
      end: 191
origin:
  kind: agent
  recorded_at: '2026-10-28T05:30:05.258Z'
id: REQ-valuation-fx-rate-lookup
type: req
---
The FX rate used for a currency on a date must be the latest rate for that currency dated on or before the date, and a missing rate or two different rates for the chosen date must be an error.

## Context

NBP publishes table A only on business days, so a weekend or holiday valuation needs the last published rate; a later rate must never leak into an earlier date. The rule is a pure function over whatever rates the swappable rate provider (D-20) supplied, so it does not depend on NBP. Polish tax uses the same lookup with an earlier cutoff (REQ-pl-tax-fx-rate-date). Refines REQ-AUTO-5769411DB9663682. Modeled by Claude Code for the M0 valuation task; the author has not reviewed it yet.

## Source

> In M0, NBP table A rates convert amounts to PLN, behind a swappable rate-provider interface with NBP as the default FX plugin.

docs/DESIGN.md D-20

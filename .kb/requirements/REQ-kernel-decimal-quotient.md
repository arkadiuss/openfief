---
title: Dividing two decimals must round the quotient exactly once, to a number of places and with a rounding mode the caller states, and dividing by zero must fail with an error.
status: open
text_ref: AGENTS.md#hard-rules
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:agents-md-hard-rules
  - lane:strict
semantic_text: Dividing two decimals must round the quotient exactly once, to a number of places and with a rounding mode the caller states, and dividing by zero must fail with an error.
logic_claims:
  - CLAIM-B0ECC6405910B17C
semantic_clauses:
  - Dividing two decimals must round the quotient exactly once, to a number of places and with a rounding mode the caller states, and dividing by zero must fail with an error.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: db7c9026ca63f97ded08e390e08b0d6e94f145d9fe52b7ef931ae85b36123cfe
semantic_inventory:
  - claim_key: CLAIM-B0ECC6405910B17C
    claim_text: Dividing two decimals must round the quotient exactly once, to a number of places and with a rounding mode the caller states, and dividing by zero must fail with an error
    role: normative
    status: modeled
    span:
      start: 0
      end: 170
origin:
  kind: agent
  recorded_at: '2026-10-28T05:11:46.612Z'
id: REQ-kernel-decimal-quotient
type: req
---
Dividing two decimals must round the quotient exactly once, to a number of places and with a rounding mode the caller states, and dividing by zero must fail with an error.

## Context

Valuation needs division: bond interest accrues over a fraction of an interest period (days elapsed over days in the period), and Polish issue letters round that result to whole grosze. Add, Sub and Mul are exact, but a quotient generally is not, so division always takes an explicit rounding; computing the numerator exactly and dividing last avoids double rounding. Extends REQ-kernel-decimal-exact (exact arithmetic, rounding only on request) to division. Modeled by Claude Code for the M0 valuation task; the author has not reviewed it yet.

## Source

> Domain logic is pure functions with table tests. No `float64` for money.

AGENTS.md#hard-rules

> Never `float64` for amounts, prices, rates or quantities.

docs/dev/code-organization.md#rules

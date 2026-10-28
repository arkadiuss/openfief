---
title: Converting an amount to PLN must multiply it exactly by the rate for its own currency and round the result once to grosze with a rounding mode the caller states, leaving the original amount unchanged.
status: open
text_ref: docs/DESIGN.md#D-20
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-20
  - lane:strict
semantic_text: Converting an amount to PLN must multiply it exactly by the rate for its own currency and round the result once to grosze with a rounding mode the caller states, leaving the original amount unchanged.
logic_claims:
  - CLAIM-5A8648671B42B58A
semantic_clauses:
  - Converting an amount to PLN must multiply it exactly by the rate for its own currency and round the result once to grosze with a rounding mode the caller states, leaving the original amount unchanged.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 476ce066b0393aa75b87e6694be8dcf7d4a561a5098ee64c7a32a120d6ca73f8
semantic_inventory:
  - claim_key: CLAIM-5A8648671B42B58A
    claim_text: Converting an amount to PLN must multiply it exactly by the rate for its own currency and round the result once to grosze with a rounding mode the caller states, leaving the original amount unchanged
    role: normative
    status: modeled
    span:
      start: 0
      end: 199
origin:
  kind: agent
  recorded_at: '2026-10-28T05:30:10.483Z'
id: REQ-valuation-pln-conversion
type: req
---
Converting an amount to PLN must multiply it exactly by the rate for its own currency and round the result once to grosze with a rounding mode the caller states, leaving the original amount unchanged.

## Context

Every amount keeps its original currency (D-20) and conversion is an explicit valuation step (glossary: Money). Tax and display may round differently, so the caller states the rounding mode. A rate for another currency is a CurrencyMismatchError, never a silent wrong number. Refines REQ-AUTO-5CB834BAA25E2C4C and REQ-AUTO-5769411DB9663682. Modeled by Claude Code for the M0 valuation task; the author has not reviewed it yet.

## Source

> Every amount keeps its original currency.

docs/DESIGN.md D-20

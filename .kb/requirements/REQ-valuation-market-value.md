---
title: A market valuation must be the holding's quantity times the instrument's end-of-day close, exactly, in the price's own currency, with the price's date and source.
status: open
text_ref: docs/DESIGN.md#D-21
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-21
  - lane:strict
semantic_text: A market valuation must be the holding's quantity times the instrument's end-of-day close, exactly, in the price's own currency, with the price's date and source.
logic_claims:
  - CLAIM-02AE57C115822AA7
semantic_clauses:
  - A market valuation must be the holding's quantity times the instrument's end-of-day close, exactly, in the price's own currency, with the price's date and source.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 9822c450a0aa0bf2d14476c0fdda86ba43bdc060463fb125856f406867e19d96
semantic_inventory:
  - claim_key: CLAIM-02AE57C115822AA7
    claim_text: A market valuation must be the holding's quantity times the instrument's end-of-day close, exactly, in the price's own currency, with the price's date and source
    role: normative
    status: modeled
    span:
      start: 0
      end: 161
origin:
  kind: agent
  recorded_at: '2026-10-28T05:30:26.654Z'
id: REQ-valuation-market-value
type: req
---
A market valuation must be the holding's quantity times the instrument's end-of-day close, exactly, in the price's own currency, with the price's date and source.

## Context

Prices are end of day in M0 (D-21) and amounts keep their original currency (D-20), so the market value stays in the quote currency; converting to PLN and rounding are separate, explicit steps. Refines REQ-AUTO-A129ED289B2239A8. Modeled by Claude Code for the M0 valuation task; the author has not reviewed it yet.

## Source

> Prices are end of day in M0.

docs/DESIGN.md D-21

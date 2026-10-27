---
title: Prices must come from a market-data connector plugin configured with the owner's own API key.
status: open
text_ref: design-decisions:D-22#1
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-22-1
  - lane:strict
semantic_text: Prices must come from a market-data connector plugin configured with the owner's own API key.
logic_claims:
  - CLAIM-66B924C649C15B1D
semantic_clauses:
  - Prices must come from a market-data connector plugin configured with the owner's own API key.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 661599e32bd14e294951c72ed7258dc3b915134936b7acf840918e3e53797914
semantic_inventory:
  - claim_key: CLAIM-66B924C649C15B1D
    claim_text: Prices must come from a market-data connector plugin configured with the owner's own API key
    role: normative
    status: modeled
    span:
      start: 0
      end: 92
origin:
  kind: import
  ref: design-decisions:D-22#1
  recorded_at: '2026-10-27T20:12:37.998Z'
firmness: firm
required_evidence: asserted
id: REQ-AUTO-0C514DFCDB36FBEB
type: req
---
Prices must come from a market-data connector plugin configured with the owner's own API key.

## Source

> Prices come from a market-data connector plugin where the user brings the API key. The vendor is chosen after checking Warsaw coverage and terms (OQ-21); IBKR Flex prices serve as a cross-check.

Source: OpenFief author decisions (DESIGN.md section 10) - D-22#1

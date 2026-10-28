---
title: A date range in which the rate provider published no table, such as a weekend or a holiday, must yield an empty set of rates, not an error.
status: open
text_ref: docs/DESIGN.md#D-20
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-20
  - lane:strict
semantic_text: A date range in which the rate provider published no table, such as a weekend or a holiday, must yield an empty set of rates, not an error.
logic_claims:
  - CLAIM-E7D2F3F724AC05C7
semantic_clauses:
  - A date range in which the rate provider published no table, such as a weekend or a holiday, must yield an empty set of rates, not an error.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 14a1cf0723c07a6d8363f3627d9965fb0609762010dd04cf91769988a614dd75
semantic_inventory:
  - claim_key: CLAIM-E7D2F3F724AC05C7
    claim_text: A date range in which the rate provider published no table, such as a weekend or a holiday, must yield an empty set of rates, not an error
    role: normative
    status: modeled
    span:
      start: 0
      end: 138
origin:
  kind: agent
  recorded_at: '2026-10-28T05:54:30.402Z'
id: REQ-valuation-fx-no-data
type: req
---
A date range in which the rate provider published no table, such as a weekend or a holiday, must yield an empty set of rates, not an error.

## Context

NBP publishes table A only on business days and answers 404 for a range without a table. Treating that as an error would fail every weekend refresh; the rate for such a date is the latest earlier one (REQ-valuation-fx-rate-lookup). Refines REQ-AUTO-5769411DB9663682 and REQ-AUTO-D1568676DF81CAF7. Modeled by Claude Code for the M0 NBP FX source task; the author has not reviewed it yet.

## Source

> In M0, NBP table A rates convert amounts to PLN, behind a swappable rate-provider interface with NBP as the default FX plugin.

docs/DESIGN.md D-20

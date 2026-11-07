---
title: The treasury bond connector must read series rates from the ministry's spreadsheet, finding its link on the gov.pl page on every fetch and falling back to dane.gov.pl's dataset 805, within size limits, and never from obligacjeskarbowe.pl.
status: open
text_ref: docs/DESIGN.md#D-97
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-97
  - lane:strict
semantic_text: The treasury bond connector must read series rates from the ministry's spreadsheet, finding its link on the gov.pl page on every fetch and falling back to dane.gov.pl's dataset 805, within size limits, and never from obligacjeskarbowe.pl.
logic_claims:
  - CLAIM-C96F2C9F19C1176B
semantic_clauses:
  - The treasury bond connector must read series rates from the ministry's spreadsheet, finding its link on the gov.pl page on every fetch and falling back to dane.gov.pl's dataset 805, within size limits, and never from obligacjeskarbowe.pl.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: e3e586b7d1be8e74acc4de24944bed126c5dcb38f9fc1cf318b05d9b78fe3079
semantic_inventory:
  - claim_key: CLAIM-C96F2C9F19C1176B
    claim_text: The treasury bond connector must read series rates from the ministry's spreadsheet, finding its link on the gov.pl page on every fetch and falling back to dane.gov.pl's dataset 805, within size limits, and never from obligacjeskarbowe.pl
    role: normative
    status: modeled
    span:
      start: 0
      end: 237
origin:
  kind: agent
  recorded_at: '2026-11-07T08:06:39.733Z'
id: REQ-valuation-mfbonds-published-rates
type: req
---
The treasury bond connector must read series rates from the ministry's spreadsheet, finding its link on the gov.pl page on every fetch and falling back to dane.gov.pl's dataset 805, within size limits, and never from obligacjeskarbowe.pl.

## Context

Author decision D-97, given after R-27 found obligacjeskarbowe.pl carries no grant of use while the ministry publishes the spreadsheet and the issue letters itself. It refines the firm D-97 requirements. Modeled by Claude Code for the D-97 bond terms work; the author has not reviewed the wording.

## Source

> The treasury bond connector reads each series' first-period rate and margin from the Ministry of Finance's published retail bond data (its spreadsheet on gov.pl, or the same dataset on dane.gov.pl), which the ministry lets anyone re-use when it is named as the source. Early-redemption fees come from a reviewed table per bond type with the dates each fee applies from, taken from the issue letters. The owner can override any series' terms, and the latest entry wins. The connector never reads obligacjeskarbowe.pl, whose pages carry no grant of use (D-8).

docs/DESIGN.md#D-97

---
title: Series rates taken from the ministry's published data must carry an attribution naming the ministry as source, the dataset and its licence, the file, when the data was made and taken, and how it was processed.
status: open
text_ref: docs/DESIGN.md#D-97
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-97
  - lane:strict
semantic_text: Series rates taken from the ministry's published data must carry an attribution naming the ministry as source, the dataset and its licence, the file, when the data was made and taken, and how it was processed.
logic_claims:
  - CLAIM-9D58DC4B640578C0
semantic_clauses:
  - Series rates taken from the ministry's published data must carry an attribution naming the ministry as source, the dataset and its licence, the file, when the data was made and taken, and how it was processed.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: d725d511373d448171b8e111816ff4f2bb79d0cbcee43aeeb036e26982f0bf99
semantic_inventory:
  - claim_key: CLAIM-9D58DC4B640578C0
    claim_text: Series rates taken from the ministry's published data must carry an attribution naming the ministry as source, the dataset and its licence, the file, when the data was made and taken, and how it was processed
    role: normative
    status: modeled
    span:
      start: 0
      end: 208
origin:
  kind: agent
  recorded_at: '2026-11-07T08:06:24.973Z'
id: REQ-pl-published-series-rates
type: req
---
Series rates taken from the ministry's published data must carry an attribution naming the ministry as source, the dataset and its licence, the file, when the data was made and taken, and how it was processed.

## Context

Author decision D-97, given after R-27 found obligacjeskarbowe.pl carries no grant of use while the ministry publishes the spreadsheet and the issue letters itself. It refines the firm D-97 requirements. The licence is CC BY 4.0 (R-27). Modeled by Claude Code for the D-97 bond terms work; the author has not reviewed the wording.

## Source

> The treasury bond connector reads each series' first-period rate and margin from the Ministry of Finance's published retail bond data (its spreadsheet on gov.pl, or the same dataset on dane.gov.pl), which the ministry lets anyone re-use when it is named as the source. Early-redemption fees come from a reviewed table per bond type with the dates each fee applies from, taken from the issue letters. The owner can override any series' terms, and the latest entry wins. The connector never reads obligacjeskarbowe.pl, whose pages carry no grant of use (D-8).

docs/DESIGN.md#D-97

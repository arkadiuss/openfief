---
title: The core must store the attribution of published series terms with them, and terms stored later for the series must replace them and their attribution.
status: open
text_ref: docs/DESIGN.md#D-97
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-97
  - lane:strict
semantic_text: The core must store the attribution of published series terms with them, and terms stored later for the series must replace them and their attribution.
logic_claims:
  - CLAIM-B42C9E0FEA89A9F9
semantic_clauses:
  - The core must store the attribution of published series terms with them, and terms stored later for the series must replace them and their attribution.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: dd70249f736c2be7c941f4dcb2f4914f58127ec6dbc6d57a7a9d0d45c22c9f87
semantic_inventory:
  - claim_key: CLAIM-B42C9E0FEA89A9F9
    claim_text: The core must store the attribution of published series terms with them, and terms stored later for the series must replace them and their attribution
    role: normative
    status: modeled
    span:
      start: 0
      end: 150
origin:
  kind: agent
  recorded_at: '2026-11-07T08:06:45.279Z'
id: REQ-valuation-bond-terms-attribution-stored
type: req
---
The core must store the attribution of published series terms with them, and terms stored later for the series must replace them and their attribution.

## Context

Author decision D-97, given after R-27 found obligacjeskarbowe.pl carries no grant of use while the ministry publishes the spreadsheet and the issue letters itself. It refines the firm D-97 requirements. Modeled by Claude Code for the D-97 bond terms work; the author has not reviewed the wording.

## Source

> The treasury bond connector reads each series' first-period rate and margin from the Ministry of Finance's published retail bond data (its spreadsheet on gov.pl, or the same dataset on dane.gov.pl), which the ministry lets anyone re-use when it is named as the source. Early-redemption fees come from a reviewed table per bond type with the dates each fee applies from, taken from the issue letters. The owner can override any series' terms, and the latest entry wins. The connector never reads obligacjeskarbowe.pl, whose pages carry no grant of use (D-8).

docs/DESIGN.md#D-97

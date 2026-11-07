---
title: The owner must be able to leave a bond lot's first-period rate and margin, both or neither, and its early-redemption fee to the connector.
status: open
text_ref: docs/DESIGN.md#D-97
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-97
  - lane:strict
semantic_text: The owner must be able to leave a bond lot's first-period rate and margin, both or neither, and its early-redemption fee to the connector.
logic_claims:
  - CLAIM-ACF05F9368602BCE
semantic_clauses:
  - The owner must be able to leave a bond lot's first-period rate and margin, both or neither, and its early-redemption fee to the connector.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 6dcc27468db126c0bb33762a9d96e1e270096a70394a47a473d1aa1fd6aa142d
semantic_inventory:
  - claim_key: CLAIM-ACF05F9368602BCE
    claim_text: The owner must be able to leave a bond lot's first-period rate and margin, both or neither, and its early-redemption fee to the connector
    role: normative
    status: modeled
    span:
      start: 0
      end: 137
origin:
  kind: agent
  recorded_at: '2026-11-07T08:06:58.434Z'
id: REQ-ingest-bond-lot-terms-optional
type: req
---
The owner must be able to leave a bond lot's first-period rate and margin, both or neither, and its early-redemption fee to the connector.

## Context

Author decision D-97, given after R-27 found obligacjeskarbowe.pl carries no grant of use while the ministry publishes the spreadsheet and the issue letters itself. It refines the firm D-97 requirements. Modeled by Claude Code for the D-97 bond terms work; the author has not reviewed the wording.

## Source

> The treasury bond connector reads each series' first-period rate and margin from the Ministry of Finance's published retail bond data (its spreadsheet on gov.pl, or the same dataset on dane.gov.pl), which the ministry lets anyone re-use when it is named as the source. Early-redemption fees come from a reviewed table per bond type with the dates each fee applies from, taken from the issue letters. The owner can override any series' terms, and the latest entry wins. The connector never reads obligacjeskarbowe.pl, whose pages carry no grant of use (D-8).

docs/DESIGN.md#D-97

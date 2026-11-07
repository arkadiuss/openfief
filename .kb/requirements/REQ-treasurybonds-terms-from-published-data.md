---
title: Terms the owner leaves out of a bond lot must be filled in by the connector from the published data and the reviewed fee table, and terms the owner enters must override them.
status: open
text_ref: docs/DESIGN.md#D-97
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-97
  - lane:strict
semantic_text: Terms the owner leaves out of a bond lot must be filled in by the connector from the published data and the reviewed fee table, and terms the owner enters must override them.
logic_claims:
  - CLAIM-F2297140CF2247EC
semantic_clauses:
  - Terms the owner leaves out of a bond lot must be filled in by the connector from the published data and the reviewed fee table, and terms the owner enters must override them.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 966649b45dcf24f68a2dfcb78d17710c88cd806d7436ae99b2137d6f80891de8
semantic_inventory:
  - claim_key: CLAIM-F2297140CF2247EC
    claim_text: Terms the owner leaves out of a bond lot must be filled in by the connector from the published data and the reviewed fee table, and terms the owner enters must override them
    role: normative
    status: modeled
    span:
      start: 0
      end: 173
origin:
  kind: agent
  recorded_at: '2026-11-07T08:06:53.025Z'
id: REQ-treasurybonds-terms-from-published-data
type: req
---
Terms the owner leaves out of a bond lot must be filled in by the connector from the published data and the reviewed fee table, and terms the owner enters must override them.

## Context

Author decision D-97, given after R-27 found obligacjeskarbowe.pl carries no grant of use while the ministry publishes the spreadsheet and the issue letters itself. It refines the firm D-97 requirements. Modeled by Claude Code for the D-97 bond terms work; the author has not reviewed the wording.

## Source

> The treasury bond connector reads each series' first-period rate and margin from the Ministry of Finance's published retail bond data (its spreadsheet on gov.pl, or the same dataset on dane.gov.pl), which the ministry lets anyone re-use when it is named as the source. Early-redemption fees come from a reviewed table per bond type with the dates each fee applies from, taken from the issue letters. The owner can override any series' terms, and the latest entry wins. The connector never reads obligacjeskarbowe.pl, whose pages carry no grant of use (D-8).

docs/DESIGN.md#D-97

---
title: A treasury bond's instrument page must show its stored terms, each marked as published, from the fee table, assumed or entered by the owner, with the attribution line the ministry's licence asks for under published ones.
status: open
text_ref: docs/DESIGN.md#D-97
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-97
  - lane:strict
semantic_text: A treasury bond's instrument page must show its stored terms, each marked as published, from the fee table, assumed or entered by the owner, with the attribution line the ministry's licence asks for under published ones.
logic_claims:
  - CLAIM-1129280BE43B34CC
semantic_clauses:
  - A treasury bond's instrument page must show its stored terms, each marked as published, from the fee table, assumed or entered by the owner, with the attribution line the ministry's licence asks for under published ones.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: fbe8c7bc0e123fef4d3fa00b1e9fe2bd87358844493119261809cf7646f4fe44
semantic_inventory:
  - claim_key: CLAIM-1129280BE43B34CC
    claim_text: A treasury bond's instrument page must show its stored terms, each marked as published, from the fee table, assumed or entered by the owner, with the attribution line the ministry's licence asks for under published ones
    role: normative
    status: modeled
    span:
      start: 0
      end: 219
origin:
  kind: agent
  recorded_at: '2026-11-07T13:27:17.220Z'
id: REQ-ui-instrument-bond-terms
type: req
---
A treasury bond's instrument page must show its stored terms, each marked as published, from the fee table, assumed or entered by the owner, with the attribution line the ministry's licence asks for under published ones.

## Context

Author decision D-97. The design session decided the D-97 follow-ups for the author: no hard cutoff for fees past the reviewed table (carry the latest reviewed fee forward, raise one attention item per held series, the owner's confirmation or override settles it; before a type's first row keep asking the owner; re-review the table each release), the series' terms shown on its instrument page and in the read API with where each came from and the published data's attribution, and a later sync filling in published terms a lot's series lacked when it was entered, never replacing stored or owner-entered ones. The line reads Source: Ministerstwo Finansów, the dataset, CC BY 4.0, made and fetched dates, processed by OpenFief (R-27). Modeled by Claude Code for the D-97 follow-ups; the author has not reviewed the wording.

## Source

> The treasury bond connector reads each series' first-period rate and margin from the Ministry of Finance's published retail bond data (its spreadsheet on gov.pl, or the same dataset on dane.gov.pl), which the ministry lets anyone re-use when it is named as the source. Early-redemption fees come from a reviewed table per bond type with the dates each fee applies from, taken from the issue letters. The owner can override any series' terms, and the latest entry wins. The connector never reads obligacjeskarbowe.pl, whose pages carry no grant of use (D-8).

docs/DESIGN.md#D-97

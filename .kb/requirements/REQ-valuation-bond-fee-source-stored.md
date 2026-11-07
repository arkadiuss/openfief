---
title: The core must store where each series' early-redemption fee came from, entered by the owner, from the fee table or assumed, with its terms, and the source must cross the connector contract both ways.
status: open
text_ref: docs/DESIGN.md#D-97
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-97
  - lane:strict
semantic_text: The core must store where each series' early-redemption fee came from, entered by the owner, from the fee table or assumed, with its terms, and the source must cross the connector contract both ways.
logic_claims:
  - CLAIM-FC3E15AFEC8F8B5F
semantic_clauses:
  - The core must store where each series' early-redemption fee came from, entered by the owner, from the fee table or assumed, with its terms, and the source must cross the connector contract both ways.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 50c270b719f9a642a26c53a0d6be4dd6a3701aeacacf9d5eb3c84a4f1943aed5
semantic_inventory:
  - claim_key: CLAIM-FC3E15AFEC8F8B5F
    claim_text: The core must store where each series' early-redemption fee came from, entered by the owner, from the fee table or assumed, with its terms, and the source must cross the connector contract both ways
    role: normative
    status: modeled
    span:
      start: 0
      end: 198
origin:
  kind: agent
  recorded_at: '2026-11-07T13:27:04.937Z'
id: REQ-valuation-bond-fee-source-stored
type: req
---
The core must store where each series' early-redemption fee came from, entered by the owner, from the fee table or assumed, with its terms, and the source must cross the connector contract both ways.

## Context

Author decision D-97. The design session decided the D-97 follow-ups for the author: no hard cutoff for fees past the reviewed table (carry the latest reviewed fee forward, raise one attention item per held series, the owner's confirmation or override settles it; before a type's first row keep asking the owner; re-review the table each release), the series' terms shown on its instrument page and in the read API with where each came from and the published data's attribution, and a later sync filling in published terms a lot's series lacked when it was entered, never replacing stored or owner-entered ones. Stored in valuation.bond_terms.fee_source by migration 0014, which marks earlier rows whose attribution says the fee came from the table. Modeled by Claude Code for the D-97 follow-ups; the author has not reviewed the wording.

## Source

> The treasury bond connector reads each series' first-period rate and margin from the Ministry of Finance's published retail bond data (its spreadsheet on gov.pl, or the same dataset on dane.gov.pl), which the ministry lets anyone re-use when it is named as the source. Early-redemption fees come from a reviewed table per bond type with the dates each fee applies from, taken from the issue letters. The owner can override any series' terms, and the latest entry wins. The connector never reads obligacjeskarbowe.pl, whose pages carry no grant of use (D-8).

docs/DESIGN.md#D-97

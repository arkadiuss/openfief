---
title: The held series whose early-redemption fee is assumed must be the ones the owner is asked to confirm, each named with its fee and the issue letter it was carried from, and the owner confirming the fee or entering another must make it the owner's.
status: open
text_ref: docs/DESIGN.md#D-97
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-97
  - lane:strict
semantic_text: The held series whose early-redemption fee is assumed must be the ones the owner is asked to confirm, each named with its fee and the issue letter it was carried from, and the owner confirming the fee or entering another must make it the owner's.
logic_claims:
  - CLAIM-9ED11854B2747FC3
semantic_clauses:
  - The held series whose early-redemption fee is assumed must be the ones the owner is asked to confirm, each named with its fee and the issue letter it was carried from, and the owner confirming the fee or entering another must make it the owner's.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: afccf75244dc4c0f8941f268a46551be0984e8baf7951ebbb0da2d5d09102687
semantic_inventory:
  - claim_key: CLAIM-9ED11854B2747FC3
    claim_text: The held series whose early-redemption fee is assumed must be the ones the owner is asked to confirm, each named with its fee and the issue letter it was carried from, and the owner confirming the fee or entering another must make it the owner's
    role: normative
    status: modeled
    span:
      start: 0
      end: 245
origin:
  kind: agent
  recorded_at: '2026-11-07T13:26:50.679Z'
id: REQ-pl-assumed-fee-confirmed-by-owner
type: req
---
The held series whose early-redemption fee is assumed must be the ones the owner is asked to confirm, each named with its fee and the issue letter it was carried from, and the owner confirming the fee or entering another must make it the owner's.

## Context

Author decision D-97. The design session decided the D-97 follow-ups for the author: no hard cutoff for fees past the reviewed table (carry the latest reviewed fee forward, raise one attention item per held series, the owner's confirmation or override settles it; before a type's first row keep asking the owner; re-review the table each release), the series' terms shown on its instrument page and in the read API with where each came from and the published data's attribution, and a later sync filling in published terms a lot's series lacked when it was entered, never replacing stored or owner-entered ones. The question reads, for example, confirm the early-redemption fee for EDO1137: 3.00 PLN assumed from the latest reviewed letter. Modeled by Claude Code for the D-97 follow-ups; the author has not reviewed the wording.

## Source

> The treasury bond connector reads each series' first-period rate and margin from the Ministry of Finance's published retail bond data (its spreadsheet on gov.pl, or the same dataset on dane.gov.pl), which the ministry lets anyone re-use when it is named as the source. Early-redemption fees come from a reviewed table per bond type with the dates each fee applies from, taken from the issue letters. The owner can override any series' terms, and the latest entry wins. The connector never reads obligacjeskarbowe.pl, whose pages carry no grant of use (D-8).

docs/DESIGN.md#D-97

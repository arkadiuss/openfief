---
title: A series' early-redemption fee must come from the reviewed table of the issue letters by bond type and sale month; a series sold after the month the table was reviewed through must carry its type's latest reviewed fee forward, marked assumed, and a series sold before its type's first row must get none.
status: open
text_ref: docs/DESIGN.md#D-97
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-97
  - lane:strict
semantic_text: A series' early-redemption fee must come from the reviewed table of the issue letters by bond type and sale month; a series sold after the month the table was reviewed through must carry its type's latest reviewed fee forward, marked assumed, and a series sold before its type's first row must get none.
logic_claims:
  - CLAIM-18E8DC0959D01D50
semantic_clauses:
  - A series' early-redemption fee must come from the reviewed table of the issue letters by bond type and sale month; a series sold after the month the table was reviewed through must carry its type's latest reviewed fee forward, marked assumed, and a series sold before its type's first row must get none.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: ba34cb53fec298062141084c4a3f0fd6d92ae8ee7c392303ade2f5dbc7da4e11
semantic_inventory:
  - claim_key: CLAIM-18E8DC0959D01D50
    claim_text: A series' early-redemption fee must come from the reviewed table of the issue letters by bond type and sale month; a series sold after the month the table was reviewed through must carry its type's latest reviewed fee forward, marked assumed, and a series sold before its type's first row must get none
    role: normative
    status: modeled
    span:
      start: 0
      end: 302
origin:
  kind: agent
  recorded_at: '2026-11-07T13:26:46.686Z'
id: REQ-pl-fee-carried-forward-assumed
type: req
---
A series' early-redemption fee must come from the reviewed table of the issue letters by bond type and sale month; a series sold after the month the table was reviewed through must carry its type's latest reviewed fee forward, marked assumed, and a series sold before its type's first row must get none.

## Context

Author decision D-97. The design session decided the D-97 follow-ups for the author: no hard cutoff for fees past the reviewed table (carry the latest reviewed fee forward, raise one attention item per held series, the owner's confirmation or override settles it; before a type's first row keep asking the owner; re-review the table each release), the series' terms shown on its instrument page and in the read API with where each came from and the published data's attribution, and a later sync filling in published terms a lot's series lacked when it was entered, never replacing stored or owner-entered ones. This replaces REQ-pl-early-redemption-fee-table, which gave no fee for a month the table was not reviewed for. Modeled by Claude Code for the D-97 follow-ups; the author has not reviewed the wording.

## Source

> The treasury bond connector reads each series' first-period rate and margin from the Ministry of Finance's published retail bond data (its spreadsheet on gov.pl, or the same dataset on dane.gov.pl), which the ministry lets anyone re-use when it is named as the source. Early-redemption fees come from a reviewed table per bond type with the dates each fee applies from, taken from the issue letters. The owner can override any series' terms, and the latest entry wins. The connector never reads obligacjeskarbowe.pl, whose pages carry no grant of use (D-8).

docs/DESIGN.md#D-97

## Why this changed

REQ-pl-early-redemption-fee-table refused a fee for a series sold after the month the table was reviewed through, so the owner had to enter it. The design session decided for the author that there is no hard cutoff: the latest reviewed fee is carried forward and the owner is asked to confirm it.

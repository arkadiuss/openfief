---
title: A lot that leaves every term to the connector, of a series the ministry's data doesn't list yet, must be booked without terms, and a later sync must fill in the terms of each held series the core stores none of once the data lists it, never replacing stored terms.
status: open
text_ref: docs/DESIGN.md#D-97
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-97
  - lane:strict
semantic_text: A lot that leaves every term to the connector, of a series the ministry's data doesn't list yet, must be booked without terms, and a later sync must fill in the terms of each held series the core stores none of once the data lists it, never replacing stored terms.
logic_claims:
  - CLAIM-819A0058C87E75E1
semantic_clauses:
  - A lot that leaves every term to the connector, of a series the ministry's data doesn't list yet, must be booked without terms, and a later sync must fill in the terms of each held series the core stores none of once the data lists it, never replacing stored terms.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 6ade54e00f8331a04f86be820984a6055a00c9f1b49af283ee4a13b7a350ada2
semantic_inventory:
  - claim_key: CLAIM-819A0058C87E75E1
    claim_text: A lot that leaves every term to the connector, of a series the ministry's data doesn't list yet, must be booked without terms, and a later sync must fill in the terms of each held series the core stores none of once the data lists it, never replacing stored terms
    role: normative
    status: modeled
    span:
      start: 0
      end: 263
origin:
  kind: agent
  recorded_at: '2026-11-07T13:27:09.129Z'
id: REQ-treasurybonds-sync-fills-missing-terms
type: req
---
A lot that leaves every term to the connector, of a series the ministry's data doesn't list yet, must be booked without terms, and a later sync must fill in the terms of each held series the core stores none of once the data lists it, never replacing stored terms.

## Context

Author decision D-97. The design session decided the D-97 follow-ups for the author: no hard cutoff for fees past the reviewed table (carry the latest reviewed fee forward, raise one attention item per held series, the owner's confirmation or override settles it; before a type's first row keep asking the owner; re-review the table each release), the series' terms shown on its instrument page and in the read API with where each came from and the published data's attribution, and a later sync filling in published terms a lot's series lacked when it was entered, never replacing stored or owner-entered ones. The spreadsheet lags a new series by days or weeks. Modeled by Claude Code for the D-97 follow-ups; the author has not reviewed the wording.

## Source

> The treasury bond connector reads each series' first-period rate and margin from the Ministry of Finance's published retail bond data (its spreadsheet on gov.pl, or the same dataset on dane.gov.pl), which the ministry lets anyone re-use when it is named as the source. Early-redemption fees come from a reviewed table per bond type with the dates each fee applies from, taken from the issue letters. The owner can override any series' terms, and the latest entry wins. The connector never reads obligacjeskarbowe.pl, whose pages carry no grant of use (D-8).

docs/DESIGN.md#D-97

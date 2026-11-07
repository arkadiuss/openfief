---
title: The owner must be able to confirm a series' assumed early-redemption fee, which revalues nothing, or override any series' fee, which revalues every account that held the series when the amount changes, and confirming a fee that is not assumed must be refused.
status: open
text_ref: docs/DESIGN.md#D-97
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-97
  - lane:strict
semantic_text: The owner must be able to confirm a series' assumed early-redemption fee, which revalues nothing, or override any series' fee, which revalues every account that held the series when the amount changes, and confirming a fee that is not assumed must be refused.
logic_claims:
  - CLAIM-3234B71D811BBE05
semantic_clauses:
  - The owner must be able to confirm a series' assumed early-redemption fee, which revalues nothing, or override any series' fee, which revalues every account that held the series when the amount changes, and confirming a fee that is not assumed must be refused.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: cb9c36a9c2bbb2ba4a8bad21543d36a826b5572f0e70fb6bc3454289877512d8
semantic_inventory:
  - claim_key: CLAIM-3234B71D811BBE05
    claim_text: The owner must be able to confirm a series' assumed early-redemption fee, which revalues nothing, or override any series' fee, which revalues every account that held the series when the amount changes, and confirming a fee that is not assumed must be refused
    role: normative
    status: modeled
    span:
      start: 0
      end: 258
origin:
  kind: agent
  recorded_at: '2026-11-07T13:27:00.563Z'
id: REQ-valuation-owner-sets-bond-fee
type: req
---
The owner must be able to confirm a series' assumed early-redemption fee, which revalues nothing, or override any series' fee, which revalues every account that held the series when the amount changes, and confirming a fee that is not assumed must be refused.

## Context

Author decision D-97. The design session decided the D-97 follow-ups for the author: no hard cutoff for fees past the reviewed table (carry the latest reviewed fee forward, raise one attention item per held series, the owner's confirmation or override settles it; before a type's first row keep asking the owner; re-review the table each release), the series' terms shown on its instrument page and in the read API with where each came from and the published data's attribution, and a later sync filling in published terms a lot's series lacked when it was entered, never replacing stored or owner-entered ones. It refines the firm REQ-ingest-bond-terms-owner-overrides for the fee alone, through the owner's SettingsService (D-79) and the instrument page. Modeled by Claude Code for the D-97 follow-ups; the author has not reviewed the wording.

## Source

> The treasury bond connector reads each series' first-period rate and margin from the Ministry of Finance's published retail bond data (its spreadsheet on gov.pl, or the same dataset on dane.gov.pl), which the ministry lets anyone re-use when it is named as the source. Early-redemption fees come from a reviewed table per bond type with the dates each fee applies from, taken from the issue letters. The owner can override any series' terms, and the latest entry wins. The connector never reads obligacjeskarbowe.pl, whose pages carry no grant of use (D-8).

docs/DESIGN.md#D-97

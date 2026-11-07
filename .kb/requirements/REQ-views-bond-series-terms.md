---
title: The read API must return a treasury bond series' stored terms with where each came from, published, the fee table, assumed or entered by the owner, and the published data's attribution.
status: open
text_ref: docs/DESIGN.md#D-97
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-97
  - lane:strict
semantic_text: The read API must return a treasury bond series' stored terms with where each came from, published, the fee table, assumed or entered by the owner, and the published data's attribution.
logic_claims:
  - CLAIM-70B143E0660FF6B7
semantic_clauses:
  - The read API must return a treasury bond series' stored terms with where each came from, published, the fee table, assumed or entered by the owner, and the published data's attribution.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 0704c05df3087d826c9869a2aeaada7f23b4fca14bdbe51295e1bf83655861f0
semantic_inventory:
  - claim_key: CLAIM-70B143E0660FF6B7
    claim_text: The read API must return a treasury bond series' stored terms with where each came from, published, the fee table, assumed or entered by the owner, and the published data's attribution
    role: normative
    status: modeled
    span:
      start: 0
      end: 184
origin:
  kind: agent
  recorded_at: '2026-11-07T13:27:13.124Z'
id: REQ-views-bond-series-terms
type: req
---
The read API must return a treasury bond series' stored terms with where each came from, published, the fee table, assumed or entered by the owner, and the published data's attribution.

## Context

Author decision D-97. The design session decided the D-97 follow-ups for the author: no hard cutoff for fees past the reviewed table (carry the latest reviewed fee forward, raise one attention item per held series, the owner's confirmation or override settles it; before a type's first row keep asking the owner; re-review the table each release), the series' terms shown on its instrument page and in the read API with where each came from and the published data's attribution, and a later sync filling in published terms a lot's series lacked when it was entered, never replacing stored or owner-entered ones. An addition to the read API (D-50): InstrumentDetail.bond_terms, over Connect and REST. Modeled by Claude Code for the D-97 follow-ups; the author has not reviewed the wording.

## Source

> The treasury bond connector reads each series' first-period rate and margin from the Ministry of Finance's published retail bond data (its spreadsheet on gov.pl, or the same dataset on dane.gov.pl), which the ministry lets anyone re-use when it is named as the source. Early-redemption fees come from a reviewed table per bond type with the dates each fee applies from, taken from the issue letters. The owner can override any series' terms, and the latest entry wins. The connector never reads obligacjeskarbowe.pl, whose pages carry no grant of use (D-8).

docs/DESIGN.md#D-97

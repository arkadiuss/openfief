---
title: The treasury bond connector must take each series' early-redemption fee from a reviewed table per bond type, taken from the issue letters, with the dates each fee applies from.
status: open
text_ref: design-decisions:D-97#2
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-97-2
  - lane:strict
semantic_text: The treasury bond connector must take each series' early-redemption fee from a reviewed table per bond type, taken from the issue letters, with the dates each fee applies from.
logic_claims:
  - CLAIM-647CE0F76DBE5AC3
semantic_clauses:
  - The treasury bond connector must take each series' early-redemption fee from a reviewed table per bond type, taken from the issue letters, with the dates each fee applies from.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: f663136375bc619439a09017dedd5fb1e493bb697836c10ce51884ced2f4cd1f
semantic_inventory:
  - claim_key: CLAIM-647CE0F76DBE5AC3
    claim_text: The treasury bond connector must take each series' early-redemption fee from a reviewed table per bond type, taken from the issue letters, with the dates each fee applies from
    role: normative
    status: modeled
    span:
      start: 0
      end: 175
firmness: firm
required_evidence: asserted
origin:
  kind: import
  ref: design-decisions:D-97#2
  recorded_at: '2026-11-07T12:37:39.250Z'
id: REQ-treasurybonds-fees-from-reviewed-table
type: req
---
The treasury bond connector must take each series' early-redemption fee from a reviewed table per bond type, taken from the issue letters, with the dates each fee applies from.

## Context

D-9 has the owner enter treasury bond lots by hand, and the bond rules need each series' first-period rate, margin and early-redemption fee to value them. Author decision D-97, as reworded when the design session answered OQ-30, says where those terms come from: the rates from the Ministry of Finance's published retail bond data, whose licence (CC BY 4.0) lets anyone re-use it when the ministry is named as the source (R-27), the fees from a reviewed table of the issue letters (R-28), and the owner's own entry over both, the latest entry winning (REQ-ingest-bond-terms-latest-entry-wins).

## Source

> Early-redemption fees come from a reviewed table per bond type with the dates each fee applies from, taken from the issue letters.

OpenFief author decisions (docs/DESIGN.md, section 3.2), design-decisions:D-97#2

## Why this changed

REQ-treasurybonds-series-terms-from-issuer was imported from D-97's first wording, which had the connector read every series term, the fee included, from obligacjeskarbowe.pl once OQ-30 confirmed that site's terms of use allowed it. OQ-30 found no grant of use on that site and an open licence on the ministry's own data (R-27), and the author reworded D-97. Its first, second and fourth sentences replace the old requirement as three firm ones: REQ-treasurybonds-rates-from-ministry-data, REQ-treasurybonds-fees-from-reviewed-table and REQ-treasurybonds-never-reads-issuer-site. This one states where the early-redemption fee comes from; the default-firmness REQ-pl-early-redemption-fee-table refines it.

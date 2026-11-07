---
title: 'The treasury bond connector must read each series'' first-period rate and margin from the Ministry of Finance''s published retail bond data: its spreadsheet on gov.pl, or the same dataset on dane.gov.pl.'
status: open
text_ref: design-decisions:D-97#1
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-97-1
  - lane:strict
semantic_text: 'The treasury bond connector must read each series'' first-period rate and margin from the Ministry of Finance''s published retail bond data: its spreadsheet on gov.pl, or the same dataset on dane.gov.pl.'
logic_claims:
  - CLAIM-EF8FB24844D20133
semantic_clauses:
  - 'The treasury bond connector must read each series'' first-period rate and margin from the Ministry of Finance''s published retail bond data: its spreadsheet on gov.pl, or the same dataset on dane.gov.pl.'
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 97d9e745b41917416613b881be35a76d500922bc59985b04b5f518db0512e06d
semantic_inventory:
  - claim_key: CLAIM-EF8FB24844D20133
    claim_text: 'The treasury bond connector must read each series'' first-period rate and margin from the Ministry of Finance''s published retail bond data: its spreadsheet on gov.pl, or the same dataset on dane.gov.pl'
    role: normative
    status: modeled
    span:
      start: 0
      end: 200
firmness: firm
required_evidence: asserted
origin:
  kind: import
  ref: design-decisions:D-97#1
  recorded_at: '2026-11-07T12:37:34.869Z'
id: REQ-treasurybonds-rates-from-ministry-data
type: req
---
The treasury bond connector must read each series' first-period rate and margin from the Ministry of Finance's published retail bond data: its spreadsheet on gov.pl, or the same dataset on dane.gov.pl.

## Context

D-9 has the owner enter treasury bond lots by hand, and the bond rules need each series' first-period rate, margin and early-redemption fee to value them. Author decision D-97, as reworded when the design session answered OQ-30, says where those terms come from: the rates from the Ministry of Finance's published retail bond data, whose licence (CC BY 4.0) lets anyone re-use it when the ministry is named as the source (R-27), the fees from a reviewed table of the issue letters (R-28), and the owner's own entry over both, the latest entry winning (REQ-ingest-bond-terms-latest-entry-wins).

## Source

> The treasury bond connector reads each series' first-period rate and margin from the Ministry of Finance's published retail bond data (its spreadsheet on gov.pl, or the same dataset on dane.gov.pl), which the ministry lets anyone re-use when it is named as the source.

OpenFief author decisions (docs/DESIGN.md, section 3.2), design-decisions:D-97#1

## Why this changed

REQ-treasurybonds-series-terms-from-issuer was imported from D-97's first wording, which had the connector read every series term, the fee included, from obligacjeskarbowe.pl once OQ-30 confirmed that site's terms of use allowed it. OQ-30 found no grant of use on that site and an open licence on the ministry's own data (R-27), and the author reworded D-97. Its first, second and fourth sentences replace the old requirement as three firm ones: REQ-treasurybonds-rates-from-ministry-data, REQ-treasurybonds-fees-from-reviewed-table and REQ-treasurybonds-never-reads-issuer-site. This one states where the first-period rate and margin come from; the default-firmness REQ-treasurybonds-terms-from-published-data, REQ-pl-published-series-rates and REQ-valuation-mfbonds-published-rates refine how.

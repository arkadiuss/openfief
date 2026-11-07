---
title: The treasury bond connector must never read obligacjeskarbowe.pl, whose pages carry no grant of use.
status: open
text_ref: design-decisions:D-97#4
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-97-4
  - lane:strict
semantic_text: The treasury bond connector must never read obligacjeskarbowe.pl, whose pages carry no grant of use.
logic_claims:
  - CLAIM-AAEA82A30D141DC4
semantic_clauses:
  - The treasury bond connector must never read obligacjeskarbowe.pl, whose pages carry no grant of use.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: fcb25e95446ee1ecd9af3a99713ba09bb5d09dc68cec12602459d31ce91bd559
semantic_inventory:
  - claim_key: CLAIM-AAEA82A30D141DC4
    claim_text: The treasury bond connector must never read obligacjeskarbowe.pl, whose pages carry no grant of use
    role: normative
    status: modeled
    span:
      start: 0
      end: 99
firmness: firm
required_evidence: asserted
origin:
  kind: import
  ref: design-decisions:D-97#4
  recorded_at: '2026-11-07T12:37:48.423Z'
id: REQ-treasurybonds-never-reads-issuer-site
type: req
---
The treasury bond connector must never read obligacjeskarbowe.pl, whose pages carry no grant of use.

## Context

D-9 has the owner enter treasury bond lots by hand, and the bond rules need each series' first-period rate, margin and early-redemption fee to value them. Author decision D-97, as reworded when the design session answered OQ-30, says where those terms come from: the rates from the Ministry of Finance's published retail bond data, whose licence (CC BY 4.0) lets anyone re-use it when the ministry is named as the source (R-27), the fees from a reviewed table of the issue letters (R-28), and the owner's own entry over both, the latest entry winning (REQ-ingest-bond-terms-latest-entry-wins).

## Source

> The connector never reads obligacjeskarbowe.pl, whose pages carry no grant of use (D-8).

OpenFief author decisions (docs/DESIGN.md, section 3.2), design-decisions:D-97#4

## Why this changed

REQ-treasurybonds-series-terms-from-issuer was imported from D-97's first wording, which had the connector read every series term, the fee included, from obligacjeskarbowe.pl once OQ-30 confirmed that site's terms of use allowed it. OQ-30 found no grant of use on that site and an open licence on the ministry's own data (R-27), and the author reworded D-97. Its first, second and fourth sentences replace the old requirement as three firm ones: REQ-treasurybonds-rates-from-ministry-data, REQ-treasurybonds-fees-from-reviewed-table and REQ-treasurybonds-never-reads-issuer-site. This one states the site the connector never reads; the default-firmness REQ-plugins-treasury-bonds-ministry-egress refines it as the plugin's egress, which names the ministry's hosts only.

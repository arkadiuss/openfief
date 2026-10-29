---
title: Once obligacjeskarbowe.pl's terms of use are confirmed to allow it, the treasury bond connector must read each series' terms, the first-period rate, margin and early-redemption fee, from that site.
status: open
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-97-1
  - lane:strict
text_ref: design-decisions:D-97#1
origin:
  kind: import
  ref: design-decisions:D-97#1
  recorded_at: '2026-10-29T09:45:06.923Z'
semantic_text: Once obligacjeskarbowe.pl's terms of use are confirmed to allow it, the treasury bond connector must read each series' terms, the first-period rate, margin and early-redemption fee, from that site.
logic_claims:
  - CLAIM-761EA92DFFCC6BA1
semantic_clauses:
  - Once obligacjeskarbowe.pl's terms of use are confirmed to allow it, the treasury bond connector must read each series' terms, the first-period rate, margin and early-redemption fee, from that site.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
firmness: firm
required_evidence: asserted
semantic_source_hash: dbebdd77136e0c907e762da17ffcda751ffd05dc8748afcac5e6dab3b4614e2c
semantic_inventory:
  - claim_key: CLAIM-761EA92DFFCC6BA1
    claim_text: Once obligacjeskarbowe.pl's terms of use are confirmed to allow it, the treasury bond connector must read each series' terms, the first-period rate, margin and early-redemption fee, from that site
    role: normative
    status: modeled
    span:
      start: 0
      end: 196
id: REQ-treasurybonds-series-terms-from-issuer
type: req
---
Once obligacjeskarbowe.pl's terms of use are confirmed to allow it, the treasury bond connector must read each series' terms, the first-period rate, margin and early-redemption fee, from that site.

## Context

D-8 ships only connectors that comply with the institution's terms, and OQ-30 asks whether obligacjeskarbowe.pl's terms of use allow a connector to read series terms from it automatically. D-97 has the treasury bond connector read each series' terms from the issuer's site once that is confirmed; until then the owner enters them with a lot (REQ-ingest-bond-terms-entered-with-lot). REQ-plugins-treasury-bonds-connector, which the author has not reviewed, limits the connector to GUS's API and NBP's archive, which stays right until OQ-30 is answered. The author recorded D-96 to D-99 after the design session that answered the implementation's questions.

## Source

> The treasury bond connector reads each series' terms (first-period rate, margin and early-redemption fee) from the issuer's site, obligacjeskarbowe.pl, once its terms of use are confirmed to allow it (OQ-30, D-8). Until then the owner enters them with a lot, and the latest entry for a series wins.

OpenFief author decisions (docs/DESIGN.md, section 3.2), design-decisions:D-97#1

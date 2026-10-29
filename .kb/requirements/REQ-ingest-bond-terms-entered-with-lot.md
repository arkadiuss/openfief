---
title: Until obligacjeskarbowe.pl's terms of use are confirmed to allow reading series terms from it, the owner must enter a series' terms with a lot.
status: open
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-97-2
  - lane:strict
text_ref: design-decisions:D-97#2
origin:
  kind: import
  ref: design-decisions:D-97#2
  recorded_at: '2026-10-29T09:45:46.028Z'
semantic_text: Until obligacjeskarbowe.pl's terms of use are confirmed to allow reading series terms from it, the owner must enter a series' terms with a lot.
logic_claims:
  - CLAIM-809BDF28023F4BEE
semantic_clauses:
  - Until obligacjeskarbowe.pl's terms of use are confirmed to allow reading series terms from it, the owner must enter a series' terms with a lot.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
firmness: firm
required_evidence: asserted
semantic_source_hash: 2928b187d77b9e6bb5973ca9c45d02a976341c6fea1318987f6a34a83967091e
semantic_inventory:
  - claim_key: CLAIM-809BDF28023F4BEE
    claim_text: Until obligacjeskarbowe.pl's terms of use are confirmed to allow reading series terms from it, the owner must enter a series' terms with a lot
    role: normative
    status: modeled
    span:
      start: 0
      end: 142
id: REQ-ingest-bond-terms-entered-with-lot
type: req
---
Until obligacjeskarbowe.pl's terms of use are confirmed to allow reading series terms from it, the owner must enter a series' terms with a lot.

## Context

D-9 has the owner enter treasury bond lots by hand, and the bond rules need each series' first-period rate, margin and early-redemption fee to value them. D-97 has the connector read those terms from the issuer's site only once OQ-30 confirms its terms of use allow it (REQ-treasurybonds-series-terms-from-issuer); until then the owner enters them with a lot, and the latest entry for a series wins (REQ-ingest-bond-terms-latest-entry-wins). The author recorded D-96 to D-99 after the design session that answered the implementation's questions.

## Source

> The treasury bond connector reads each series' terms (first-period rate, margin and early-redemption fee) from the issuer's site, obligacjeskarbowe.pl, once its terms of use are confirmed to allow it (OQ-30, D-8). Until then the owner enters them with a lot, and the latest entry for a series wins.

OpenFief author decisions (docs/DESIGN.md, section 3.2), design-decisions:D-97#2

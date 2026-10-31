---
title: The 0.10 PLN discount of a bond bought by exchange must be taxed with the interest on redemption.
status: closed
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-100-2
  - lane:strict
text_ref: design-decisions:D-100#2
origin:
  kind: import
  ref: design-decisions:D-100#2
  recorded_at: '2026-10-29T09:55:56.244Z'
semantic_text: The 0.10 PLN discount of a bond bought by exchange must be taxed with the interest on redemption.
logic_claims:
  - CLAIM-C8FB46EFECC8AFE8
semantic_clauses:
  - The 0.10 PLN discount of a bond bought by exchange must be taxed with the interest on redemption.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
firmness: firm
required_evidence: asserted
semantic_source_hash: d31346aa58d32a9c5f0abb8709262b91332c8571cb5be84efe4562e5c74a4ad1
semantic_inventory:
  - claim_key: CLAIM-C8FB46EFECC8AFE8
    claim_text: The 0.10 PLN discount of a bond bought by exchange must be taxed with the interest on redemption
    role: normative
    status: modeled
    span:
      start: 0
      end: 96
id: REQ-pl-exchange-discount-taxed
type: req
---
The 0.10 PLN discount of a bond bought by exchange must be taxed with the interest on redemption.

## Context

D-100 lets a manual bond lot be an exchange of maturing bonds at 99.90 PLN instead of 100 PLN (REQ-ingest-bond-lot-purchase-kinds). The Polish jurisdiction pack must tax that 0.10 PLN discount together with the interest when the bond is redeemed. The author recorded D-100 to D-104 after the design session that answered the implementation's questions.

## Source

> Manual bond lots are cash purchases at 100 PLN or exchanges of maturing bonds at 99.90 PLN, whose 0.10 PLN discount is taxed with the interest on redemption. The owner can delete a lot as a correction, which supersedes it (D-18) and erases nothing.

OpenFief author decisions (docs/DESIGN.md, section 3.2), design-decisions:D-100#2

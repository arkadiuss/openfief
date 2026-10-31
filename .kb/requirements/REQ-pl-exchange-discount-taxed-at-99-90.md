---
title: The 0.10 PLN discount of a bond bought by exchange at 99.90 PLN must be taxed with the interest on redemption.
status: open
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
  recorded_at: '2026-10-31T12:05:38.000Z'
semantic_text: The 0.10 PLN discount of a bond bought by exchange at 99.90 PLN must be taxed with the interest on redemption.
logic_claims:
  - CLAIM-474683B9829DD94D
semantic_clauses:
  - The 0.10 PLN discount of a bond bought by exchange at 99.90 PLN must be taxed with the interest on redemption.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
firmness: firm
required_evidence: asserted
semantic_source_hash: f2ab98e6179d32c71523b0af3d7d2bcf6cf1702f454d7a2ae1211a19013623f8
semantic_inventory:
  - claim_key: CLAIM-474683B9829DD94D
    claim_text: The 0.10 PLN discount of a bond bought by exchange at 99.90 PLN must be taxed with the interest on redemption
    role: normative
    status: modeled
    span:
      start: 0
      end: 109
id: REQ-pl-exchange-discount-taxed-at-99-90
type: req
---
The 0.10 PLN discount of a bond bought by exchange at 99.90 PLN must be taxed with the interest on redemption.

## Why this changed

REQ-pl-exchange-discount-taxed taxed the 0.10 PLN discount of every bond bought by exchange, because D-100 priced every exchange at 99.90 PLN. On 2026-10-31 the author amended D-100 so the exchange price is the one the series' issue letter sets, and the OTS letter sets 100.00 PLN, so an OTS bought by exchange has no discount to tax (question #20 on docs/implementation-questions.md).

## Context

D-100 lets a manual bond lot be an exchange of maturing bonds at its issue letter's exchange price (REQ-ingest-bond-lot-exchange-price-by-letter). Where that price is 99.90 PLN, the Polish jurisdiction pack must tax the 0.10 PLN discount together with the interest when the bond is redeemed. The author decided this on 2026-10-31 while answering the implementation's questions.

## Source

> Manual bond lots are cash purchases at 100 PLN or exchanges of maturing bonds at the exchange price the series' issue letter sets: 99.90 PLN, whose 0.10 PLN discount is taxed with the interest on redemption, or 100.00 PLN for OTS. The owner can delete a lot as a correction, which supersedes it (D-18) and erases nothing.

OpenFief author decisions (docs/DESIGN.md, section 3.2), design-decisions:D-100#2

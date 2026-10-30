---
title: A treasury bond lot entered by hand must be refused unless it has a purchase date no later than today, a margin of zero or more, a purchase for cash at the nominal or by exchange at the exchange price, and can be valued on its purchase date by the Polish bond rules at that price.
status: open
text_ref: docs/DESIGN.md#D-100
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-100
  - lane:strict
semantic_text: A treasury bond lot entered by hand must be refused unless it has a purchase date no later than today, a margin of zero or more, a purchase for cash at the nominal or by exchange at the exchange price, and can be valued on its purchase date by the Polish bond rules at that price.
logic_claims:
  - CLAIM-B2DC2C0B0656E67B
semantic_clauses:
  - A treasury bond lot entered by hand must be refused unless it has a purchase date no later than today, a margin of zero or more, a purchase for cash at the nominal or by exchange at the exchange price, and can be valued on its purchase date by the Polish bond rules at that price.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 685e0c53abf24107587186c21bd762c5ba2c4793a0138fa75f35ed93bad0b379
semantic_inventory:
  - claim_key: CLAIM-B2DC2C0B0656E67B
    claim_text: A treasury bond lot entered by hand must be refused unless it has a purchase date no later than today, a margin of zero or more, a purchase for cash at the nominal or by exchange at the exchange price, and can be valued on its purchase date by the Polish bond rules at that price
    role: exception
    status: modeled
    span:
      start: 0
      end: 279
origin:
  kind: agent
  recorded_at: '2026-10-29T01:14:12.885Z'
id: REQ-ingest-bond-lot-rules
type: req
---
A treasury bond lot entered by hand must be refused unless it has a purchase date no later than today, a margin of zero or more, a purchase for cash at the nominal or by exchange at the exchange price, and can be valued on its purchase date by the Polish bond rules at that price.

## Context

D-9 makes manual entry the bonds' path; D-100 adds the exchange purchase at 99.90 PLN, refined here from the earlier rule that took cash purchases only. OTS keeps its letter's 100.00 PLN exchange price. Recorded by the bonds implementation agent (impl/bonds-2).

## Source

> Manual bond lots are cash purchases at 100 PLN or exchanges of maturing bonds at 99.90 PLN, whose 0.10 PLN discount is taxed with the interest on redemption. The owner can delete a lot as a correction, which supersedes it (D-18) and erases nothing.

docs/DESIGN.md#D-100

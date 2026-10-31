---
title: A manual bond lot must be either a cash purchase at 100 PLN per bond or an exchange of maturing bonds at 99.90 PLN per bond.
status: closed
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-100-1
  - lane:strict
text_ref: design-decisions:D-100#1
origin:
  kind: import
  ref: design-decisions:D-100#1
  recorded_at: '2026-10-29T09:55:10.953Z'
semantic_text: A manual bond lot must be either a cash purchase at 100 PLN per bond or an exchange of maturing bonds at 99.90 PLN per bond.
logic_claims:
  - CLAIM-7D367AC2D97EDD02
semantic_clauses:
  - A manual bond lot must be either a cash purchase at 100 PLN per bond or an exchange of maturing bonds at 99.90 PLN per bond.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
firmness: firm
required_evidence: asserted
semantic_source_hash: ef69f4f29245433af97821ede577afa13c3d2c444cc1c9e4f0948bce5dc095b5
semantic_inventory:
  - claim_key: CLAIM-7D367AC2D97EDD02
    claim_text: A manual bond lot must be either a cash purchase at 100 PLN per bond or an exchange of maturing bonds at 99.90 PLN per bond
    role: normative
    status: modeled
    span:
      start: 0
      end: 123
id: REQ-ingest-bond-lot-purchase-kinds
type: req
---
A manual bond lot must be either a cash purchase at 100 PLN per bond or an exchange of maturing bonds at 99.90 PLN per bond.

## Context

D-9 brings treasury bonds in through manual lot entry, and D-98 decides how such a lot enters the ledger. D-100 fixes the two ways a manual lot is bought: for cash at 100 PLN, or by exchanging maturing bonds at 99.90 PLN. The author recorded D-100 to D-104 after the design session that answered the implementation's questions.

## Source

> Manual bond lots are cash purchases at 100 PLN or exchanges of maturing bonds at 99.90 PLN, whose 0.10 PLN discount is taxed with the interest on redemption. The owner can delete a lot as a correction, which supersedes it (D-18) and erases nothing.

OpenFief author decisions (docs/DESIGN.md, section 3.2), design-decisions:D-100#1

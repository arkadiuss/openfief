---
title: 'A manual bond lot must be either a cash purchase at 100 PLN per bond or an exchange of maturing bonds at the exchange price its series'' issue letter sets: 99.90 PLN per bond, or 100.00 PLN per bond for OTS.'
status: open
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
  recorded_at: '2026-10-31T12:05:38.000Z'
semantic_text: 'A manual bond lot must be either a cash purchase at 100 PLN per bond or an exchange of maturing bonds at the exchange price its series'' issue letter sets: 99.90 PLN per bond, or 100.00 PLN per bond for OTS.'
logic_claims:
  - CLAIM-05F5E88D80E3AC36
semantic_clauses:
  - 'A manual bond lot must be either a cash purchase at 100 PLN per bond or an exchange of maturing bonds at the exchange price its series'' issue letter sets: 99.90 PLN per bond, or 100.00 PLN per bond for OTS.'
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
firmness: firm
required_evidence: asserted
semantic_source_hash: e817ff47a510f9ab13870b7df5750b80122ed9cd0d823f3f58eac01cf5102945
semantic_inventory:
  - claim_key: CLAIM-05F5E88D80E3AC36
    claim_text: 'A manual bond lot must be either a cash purchase at 100 PLN per bond or an exchange of maturing bonds at the exchange price its series'' issue letter sets: 99.90 PLN per bond, or 100.00 PLN per bond for OTS'
    role: normative
    status: modeled
    span:
      start: 0
      end: 205
id: REQ-ingest-bond-lot-exchange-price-by-letter
type: req
---
A manual bond lot must be either a cash purchase at 100 PLN per bond or an exchange of maturing bonds at the exchange price its series' issue letter sets: 99.90 PLN per bond, or 100.00 PLN per bond for OTS.

## Why this changed

REQ-ingest-bond-lot-purchase-kinds priced every exchange of maturing bonds at 99.90 PLN, but the OTS issue letter (ust. 8) sets its exchange price at 100.00 PLN, and the implementation already followed the letters (question #20 on docs/implementation-questions.md). On 2026-10-31 the author amended D-100 so the exchange price is the one the series' issue letter sets: 99.90 PLN, or 100.00 PLN for OTS.

## Context

D-9 brings treasury bonds in through manual lot entry, and D-98 decides how such a lot enters the ledger. D-100 fixes the two ways a manual lot is bought: for cash at 100 PLN, or by exchanging maturing bonds at the exchange price the series' issue letter sets. Only the 99.90 PLN exchange carries a 0.10 PLN discount, taxed with the interest on redemption (REQ-pl-exchange-discount-taxed-at-99-90). The author decided this on 2026-10-31 while answering the implementation's questions.

## Source

> Manual bond lots are cash purchases at 100 PLN or exchanges of maturing bonds at the exchange price the series' issue letter sets: 99.90 PLN, whose 0.10 PLN discount is taxed with the interest on redemption, or 100.00 PLN for OTS. The owner can delete a lot as a correction, which supersedes it (D-18) and erases nothing.

OpenFief author decisions (docs/DESIGN.md, section 3.2), design-decisions:D-100#1

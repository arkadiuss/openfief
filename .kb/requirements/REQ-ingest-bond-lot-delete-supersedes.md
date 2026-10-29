---
title: The owner must be able to delete a manual bond lot as a correction, which supersedes it and erases nothing.
status: open
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-100-3
  - lane:strict
text_ref: design-decisions:D-100#3
origin:
  kind: import
  ref: design-decisions:D-100#3
  recorded_at: '2026-10-29T09:56:36.711Z'
semantic_text: The owner must be able to delete a manual bond lot as a correction, which supersedes it and erases nothing.
logic_claims:
  - CLAIM-D634EA1A2EAAA4B5
semantic_clauses:
  - The owner must be able to delete a manual bond lot as a correction, which supersedes it and erases nothing.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
firmness: firm
required_evidence: asserted
semantic_source_hash: 0c66c5d3e23f81f78fe05420edb2fec41630d39d769e8831b7b21a9175735184
semantic_inventory:
  - claim_key: CLAIM-D634EA1A2EAAA4B5
    claim_text: The owner must be able to delete a manual bond lot as a correction, which supersedes it and erases nothing
    role: normative
    status: modeled
    span:
      start: 0
      end: 106
id: REQ-ingest-bond-lot-delete-supersedes
type: req
---
The owner must be able to delete a manual bond lot as a correction, which supersedes it and erases nothing.

## Context

D-18 makes the transaction log append-only and versioned: a changed row supersedes the old one, which is never deleted. D-100 applies that to a manual bond lot the owner deletes as a correction. The author recorded D-100 to D-104 after the design session that answered the implementation's questions.

## Source

> Manual bond lots are cash purchases at 100 PLN or exchanges of maturing bonds at 99.90 PLN, whose 0.10 PLN discount is taxed with the interest on redemption. The owner can delete a lot as a correction, which supersedes it (D-18) and erases nothing.

OpenFief author decisions (docs/DESIGN.md, section 3.2), design-decisions:D-100#3

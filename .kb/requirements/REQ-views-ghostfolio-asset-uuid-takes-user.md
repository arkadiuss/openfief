---
title: The UUID symbols derived for bond and fee assets must also take the Ghostfolio user's ID.
status: open
text_ref: design-decisions:D-134#1
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-134-1
  - lane:strict
semantic_text: The UUID symbols derived for bond and fee assets must also take the Ghostfolio user's ID.
logic_claims:
  - CLAIM-5A0CC9B699CE64C9
semantic_clauses:
  - The UUID symbols derived for bond and fee assets must also take the Ghostfolio user's ID.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 10d1766077623b2ca2eb5d8bc28f8f6354b5f849bd9808d109f5cec898a82351
semantic_inventory:
  - claim_key: CLAIM-5A0CC9B699CE64C9
    claim_text: The UUID symbols derived for bond and fee assets must also take the Ghostfolio user's ID
    role: normative
    status: modeled
    span:
      start: 0
      end: 88
firmness: firm
required_evidence: asserted
origin:
  kind: import
  ref: design-decisions:D-134#1
  recorded_at: '2026-11-01T07:39:14.745Z'
id: REQ-views-ghostfolio-asset-uuid-takes-user
type: req
---
The UUID symbols derived for bond and fee assets must also take the Ghostfolio user's ID.

## Context

The author recorded D-134 on 2026-10-31, extending the asset symbols of D-64 (bonds) and D-108 (fees). The decision gives the reason: two installs whose instances share a name never collide in one Ghostfolio. No end-to-end test proves it yet, so its evidence floor is asserted.

## Source

> The UUID symbols D-64 and D-108 derive for bond and fee assets also take the Ghostfolio user's ID, so two installs whose instances share a name never collide in one Ghostfolio.

OpenFief author decisions (docs/DESIGN.md, section 3.1), design-decisions:D-134#1

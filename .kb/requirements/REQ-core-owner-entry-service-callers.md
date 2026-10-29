---
title: Each owner entry service must be owner-only, versioned on its own, and called only by a view plugin acting for an owner session.
status: open
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-79-2
  - lane:strict
text_ref: design-decisions:D-79#2
origin:
  kind: import
  ref: design-decisions:D-79#2
  recorded_at: '2026-10-29T08:08:58.036Z'
semantic_text: Each owner entry service must be owner-only, versioned on its own, and called only by a view plugin acting for an owner session.
logic_claims:
  - CLAIM-539359DE04D7809D
semantic_clauses:
  - Each owner entry service must be owner-only, versioned on its own, and called only by a view plugin acting for an owner session.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
firmness: firm
required_evidence: asserted
semantic_source_hash: 97dd8c54371592cab981ec63fcef87c8a33e860392d22c6269b3a85b57b3f987
semantic_inventory:
  - claim_key: CLAIM-539359DE04D7809D
    claim_text: Each owner entry service must be owner-only, versioned on its own, and called only by a view plugin acting for an owner session
    role: normative
    status: modeled
    span:
      start: 0
      end: 127
id: REQ-core-owner-entry-service-callers
type: req
---
Each owner entry service must be owner-only, versioned on its own, and called only by a view plugin acting for an owner session.

## Context

D-79 gives each kind of hand-entered data its own service in the core (REQ-core-owner-entry-services). D-69 makes the built-in UI a backend-for-frontend that calls the core for the owner's session (REQ-ui-backend-for-frontend), and D-66 gives every session a read or owner scope. The implementation's ModelPortfolioService and BondLotService already accept only view plugins, through the AuthService's caller check, with an owner-scope session in each call (REQ-portfolio-model-entry, REQ-ingest-bond-lot-service), a shape their contexts left for the author to confirm; D-79 confirms it for every kind of entry. The author recorded D-79 and D-80 after the design session that answered the implementation's questions.

## Source

> What the owner enters by hand reaches the core through one service per kind of entry: bond lots, the model portfolio, accounts, statement uploads and any later kind. Each is owner-only, versioned on its own, and called only by a view plugin acting for an owner session (D-69). The read API (D-6) and script tokens (D-27) stay read-only.

OpenFief author decisions (docs/DESIGN.md, section 3.6), design-decisions:D-79#2

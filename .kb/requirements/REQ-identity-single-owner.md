---
title: OpenFief must have a single user, the owner, with no other users and no tenants.
status: open
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-65-1
  - lane:strict
text_ref: design-decisions:D-65#1
origin:
  kind: import
  ref: design-decisions:D-65#1
  recorded_at: '2026-10-29T05:25:26.216Z'
semantic_text: OpenFief must have a single user, the owner, with no other users and no tenants.
logic_claims:
  - CLAIM-8721AC7BF31BFCB5
semantic_clauses:
  - OpenFief must have a single user, the owner, with no other users and no tenants.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
firmness: firm
required_evidence: asserted
semantic_source_hash: 5d6712141fe37db46ac077102852f3633dd68cd9d206a70ad8bfd77f250e4d24
semantic_inventory:
  - claim_key: CLAIM-8721AC7BF31BFCB5
    claim_text: OpenFief must have a single user, the owner, with no other users and no tenants
    role: normative
    status: modeled
    span:
      start: 0
      end: 79
id: REQ-identity-single-owner
type: req
---
OpenFief must have a single user, the owner, with no other users and no tenants.

## Context

GR-12 puts the security model first, per use case. M0 has one use case, tracking and imports; the use cases of later milestones are in the proposal files. The author recorded this decision with D-66 to D-72 after the design session that answered the implementation's questions.

## Source

> OpenFief has a single user, the owner: no other users and no tenants. Authenticators only prove that a caller is the owner, and they are separate from the session that follows. In M0 the authenticators are passkeys (D-26) and, on a demo, D-68's demo authenticator.

OpenFief author decisions (docs/DESIGN.md, section 3.7), design-decisions:D-65#1

---
title: Authenticators must only prove that a caller is the owner and must be separate from the session that follows; in M0 the authenticators must be passkeys and, on a demo, the demo authenticator.
status: open
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-65-2
  - lane:strict
text_ref: design-decisions:D-65#2
origin:
  kind: import
  ref: design-decisions:D-65#2
  recorded_at: '2026-10-29T05:26:19.419Z'
semantic_text: Authenticators must only prove that a caller is the owner and must be separate from the session that follows; in M0 the authenticators must be passkeys and, on a demo, the demo authenticator.
logic_claims:
  - CLAIM-9532B3D91AA7A860
semantic_clauses:
  - Authenticators must only prove that a caller is the owner and must be separate from the session that follows; in M0 the authenticators must be passkeys and, on a demo, the demo authenticator.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
firmness: firm
required_evidence: asserted
semantic_source_hash: 0be9f21ee69d0428c4f51f614ff42fb947fc796f4bf213eda1856f4bbac1383b
semantic_inventory:
  - claim_key: CLAIM-9532B3D91AA7A860
    claim_text: Authenticators must only prove that a caller is the owner and must be separate from the session that follows; in M0 the authenticators must be passkeys and, on a demo, the demo authenticator
    role: normative
    status: modeled
    span:
      start: 0
      end: 190
id: REQ-identity-authenticators-prove-owner
type: req
---
Authenticators must only prove that a caller is the owner and must be separate from the session that follows; in M0 the authenticators must be passkeys and, on a demo, the demo authenticator.

## Context

GR-12 puts the security model first, per use case. M0 has one use case, tracking and imports; the use cases of later milestones are in the proposal files. The author recorded this decision with D-66 to D-72 after the design session that answered the implementation's questions. D-26 makes passkeys the login, and D-68 defines the demo authenticator.

## Source

> OpenFief has a single user, the owner: no other users and no tenants. Authenticators only prove that a caller is the owner, and they are separate from the session that follows. In M0 the authenticators are passkeys (D-26) and, on a demo, D-68's demo authenticator.

OpenFief author decisions (docs/DESIGN.md, section 3.7), design-decisions:D-65#2

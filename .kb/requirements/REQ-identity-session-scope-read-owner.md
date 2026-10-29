---
title: Each session must have a scope, read or owner.
status: open
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-66-2
  - lane:strict
text_ref: design-decisions:D-66#2
origin:
  kind: import
  ref: design-decisions:D-66#2
  recorded_at: '2026-10-29T05:27:47.708Z'
semantic_text: Each session must have a scope, read or owner.
logic_claims:
  - CLAIM-4E4BC215B8564B9C
semantic_clauses:
  - Each session must have a scope, read or owner.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
firmness: firm
required_evidence: asserted
semantic_source_hash: ab24a6b1892979859a91ac4a93387b7a7a1da2730f0ec6d6626c79942152f746
semantic_inventory:
  - claim_key: CLAIM-4E4BC215B8564B9C
    claim_text: Each session must have a scope, read or owner
    role: normative
    status: modeled
    span:
      start: 0
      end: 45
id: REQ-identity-session-scope-read-owner
type: req
---
Each session must have a scope, read or owner.

## Context

GR-12 puts the security model first, per use case. M0 has one use case, tracking and imports; the use cases of later milestones are in the proposal files. The author recorded this decision with D-65 and D-67 to D-72 after the design session that answered the implementation's questions. REQ-identity-session-scopes refines it for the implementation, naming what each scope may call; that list is the implementation's choice, not the author's.

## Source

> After authentication the core issues an opaque, revocable session token, not a JSON Web Token (JWT). It stores only the token's hash, and the session expires after a period of inactivity and after an absolute lifetime. Each session has a scope: `read` or `owner`.

OpenFief author decisions (docs/DESIGN.md, section 3.7), design-decisions:D-66#2

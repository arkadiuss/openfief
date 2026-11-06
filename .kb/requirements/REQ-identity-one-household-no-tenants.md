---
title: One OpenFief must serve one household, with no tenants.
status: open
text_ref: design-decisions:D-65#1
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-65-1
  - lane:strict
semantic_text: One OpenFief must serve one household, with no tenants.
logic_claims:
  - CLAIM-C53584948BA1D038
semantic_clauses:
  - One OpenFief must serve one household, with no tenants.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 2db8bdf523d492d57c2cb16c1f5354a74887cc6f2e2f56ffc732f7eb16cb2840
semantic_inventory:
  - claim_key: CLAIM-C53584948BA1D038
    claim_text: One OpenFief must serve one household, with no tenants
    role: normative
    status: modeled
    span:
      start: 0
      end: 54
firmness: firm
required_evidence: asserted
origin:
  kind: import
  ref: design-decisions:D-65#1
  recorded_at: '2026-11-02T00:52:59.460Z'
id: REQ-identity-one-household-no-tenants
type: req
---
One OpenFief must serve one household, with no tenants.

## Why this changed

On 2026-11-01 the author rewrote D-65 while planning viewers for the second alpha (reason not stated): one OpenFief serves one household, and the owner can enroll viewers. REQ-identity-single-owner said OpenFief has a single user, the owner, with no other users, which viewers contradict, so this requirement replaces it, and it is closed.

## Context

The author recorded it on 2026-11-01, when planning viewers for the second alpha. Reason not stated. GR-2 has one person own the installation; D-65 lets that owner share it with viewers in the same household. No end-to-end test proves it yet, so its evidence floor is asserted.

## Source

> One OpenFief serves one household, with no tenants. The owner holds every account and alone changes anything; from the second alpha the owner can enroll viewers, each with their own passkey, and grant each the accounts they may view. Authenticators only prove who a caller is, and they are separate from the session that follows. In M0 the authenticators are passkeys (D-26) and, on a demo, D-68's demo authenticator.

OpenFief author decisions (docs/DESIGN.md, section 3.7), design-decisions:D-65#1

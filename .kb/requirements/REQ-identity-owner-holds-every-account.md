---
title: The owner must hold every account and alone change anything.
status: open
text_ref: design-decisions:D-65#3
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-65-3
  - lane:strict
semantic_text: The owner must hold every account and alone change anything.
logic_claims:
  - CLAIM-E239B6DC0CB51CE0
semantic_clauses:
  - The owner must hold every account and alone change anything.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 31a0a9890767e0ef8b4c6fa5224ca99182295b5b81f53f410e0a974a98f68e02
semantic_inventory:
  - claim_key: CLAIM-E239B6DC0CB51CE0
    claim_text: The owner must hold every account and alone change anything
    role: normative
    status: modeled
    span:
      start: 0
      end: 59
firmness: firm
required_evidence: asserted
origin:
  kind: import
  ref: design-decisions:D-65#3
  recorded_at: '2026-11-02T00:53:14.564Z'
id: REQ-identity-owner-holds-every-account
type: req
---
The owner must hold every account and alone change anything.

## Context

The author recorded it on 2026-11-01, when planning viewers for the second alpha. Reason not stated. Viewers (D-65) only view the accounts the owner grants them; D-163's authorization interface lets the owner act on every account. No end-to-end test proves it yet, so its evidence floor is asserted.

## Source

> One OpenFief serves one household, with no tenants. The owner holds every account and alone changes anything; from the second alpha the owner can enroll viewers, each with their own passkey, and grant each the accounts they may view. Authenticators only prove who a caller is, and they are separate from the session that follows. In M0 the authenticators are passkeys (D-26) and, on a demo, D-68's demo authenticator.

OpenFief author decisions (docs/DESIGN.md, section 3.7), design-decisions:D-65#3

---
title: From the second alpha, the owner must be able to enroll viewers, each with their own passkey, and grant each the accounts they may view.
status: open
text_ref: design-decisions:D-65#4
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-65-4
  - lane:strict
semantic_text: From the second alpha, the owner must be able to enroll viewers, each with their own passkey, and grant each the accounts they may view.
logic_claims:
  - CLAIM-BDE9F75175D79085
semantic_clauses:
  - From the second alpha, the owner must be able to enroll viewers, each with their own passkey, and grant each the accounts they may view.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 72e9789c5dd6db960758820c901ff5d0406e281942b50278b6b7c15370dc5471
semantic_inventory:
  - claim_key: CLAIM-BDE9F75175D79085
    claim_text: From the second alpha, the owner must be able to enroll viewers, each with their own passkey, and grant each the accounts they may view
    role: normative
    status: modeled
    span:
      start: 0
      end: 135
firmness: firm
required_evidence: asserted
origin:
  kind: import
  ref: design-decisions:D-65#4
  recorded_at: '2026-11-02T00:53:24.669Z'
id: REQ-identity-owner-enrolls-viewers
type: req
---
From the second alpha, the owner must be able to enroll viewers, each with their own passkey, and grant each the accounts they may view.

## Why this changed

On 2026-11-01 the author rewrote D-65 while planning viewers for the second alpha (reason not stated): one OpenFief serves one household, and the owner can enroll viewers. REQ-identity-single-owner said OpenFief has a single user, the owner, with no other users, which viewers contradict, so this requirement replaces it, and it is closed.

## Context

The author recorded it on 2026-11-01, when planning viewers for the second alpha. Reason not stated. D-164 has SpiceDB answer which accounts a viewer may view from the second alpha. Not built: the first alpha has no viewers. No end-to-end test proves it yet, so its evidence floor is asserted.

## Source

> One OpenFief serves one household, with no tenants. The owner holds every account and alone changes anything; from the second alpha the owner can enroll viewers, each with their own passkey, and grant each the accounts they may view. Authenticators only prove who a caller is, and they are separate from the session that follows. In M0 the authenticators are passkeys (D-26) and, on a demo, D-68's demo authenticator.

OpenFief author decisions (docs/DESIGN.md, section 3.7), design-decisions:D-65#4

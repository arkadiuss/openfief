---
title: Authenticators must only prove who a caller is and must be separate from the session that follows; in M0 the authenticators must be passkeys and, on a demo, the demo authenticator.
status: open
text_ref: design-decisions:D-65#2
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-65-2
  - lane:strict
semantic_text: Authenticators must only prove who a caller is and must be separate from the session that follows; in M0 the authenticators must be passkeys and, on a demo, the demo authenticator.
logic_claims:
  - CLAIM-364B95027CE9FFE9
semantic_clauses:
  - Authenticators must only prove who a caller is and must be separate from the session that follows; in M0 the authenticators must be passkeys and, on a demo, the demo authenticator.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 83be36c433ccf6d6dbea7e31939f8aab46a7d74494843c4e6d706898e9130aca
semantic_inventory:
  - claim_key: CLAIM-364B95027CE9FFE9
    claim_text: Authenticators must only prove who a caller is and must be separate from the session that follows; in M0 the authenticators must be passkeys and, on a demo, the demo authenticator
    role: normative
    status: modeled
    span:
      start: 0
      end: 179
firmness: firm
required_evidence: asserted
origin:
  kind: import
  ref: design-decisions:D-65#2
  recorded_at: '2026-11-02T00:53:05.518Z'
id: REQ-identity-authenticators-prove-caller
type: req
---
Authenticators must only prove who a caller is and must be separate from the session that follows; in M0 the authenticators must be passkeys and, on a demo, the demo authenticator.

## Why this changed

On 2026-11-01 the author rewrote D-65 while planning viewers for the second alpha (reason not stated). REQ-identity-authenticators-prove-owner had authenticators prove that a caller is the owner; with viewers they prove who a caller is, so this requirement replaces it, and it is closed.

## Context

The author recorded it on 2026-11-01, when planning viewers for the second alpha. Reason not stated. Passkeys are D-26's; the demo authenticator is D-68's; the session that follows is D-66's opaque, scoped session. No end-to-end test proves it yet, so its evidence floor is asserted.

## Source

> One OpenFief serves one household, with no tenants. The owner holds every account and alone changes anything; from the second alpha the owner can enroll viewers, each with their own passkey, and grant each the accounts they may view. Authenticators only prove who a caller is, and they are separate from the session that follows. In M0 the authenticators are passkeys (D-26) and, on a demo, D-68's demo authenticator.

OpenFief author decisions (docs/DESIGN.md, section 3.7), design-decisions:D-65#2

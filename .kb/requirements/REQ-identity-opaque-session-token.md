---
title: After authentication the core must issue an opaque, revocable session token, not a JSON Web Token (JWT), store only the token's hash, and expire the session after a period of inactivity and after an absolute lifetime.
status: open
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-66-1
  - lane:strict
text_ref: design-decisions:D-66#1
origin:
  kind: import
  ref: design-decisions:D-66#1
  recorded_at: '2026-10-29T05:27:05.340Z'
semantic_text: After authentication the core must issue an opaque, revocable session token, not a JSON Web Token (JWT), store only the token's hash, and expire the session after a period of inactivity and after an absolute lifetime.
logic_claims:
  - CLAIM-56E2628A4496EEF8
semantic_clauses:
  - After authentication the core must issue an opaque, revocable session token, not a JSON Web Token (JWT), store only the token's hash, and expire the session after a period of inactivity and after an absolute lifetime.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
firmness: firm
required_evidence: asserted
semantic_source_hash: 151480b75427e0c21055f31130433618921a5250118f8adcf5738cfe6c484f35
semantic_inventory:
  - claim_key: CLAIM-56E2628A4496EEF8
    claim_text: After authentication the core must issue an opaque, revocable session token, not a JSON Web Token (JWT), store only the token's hash, and expire the session after a period of inactivity and after an absolute lifetime
    role: normative
    status: modeled
    span:
      start: 0
      end: 216
id: REQ-identity-opaque-session-token
type: req
---
After authentication the core must issue an opaque, revocable session token, not a JSON Web Token (JWT), store only the token's hash, and expire the session after a period of inactivity and after an absolute lifetime.

## Context

GR-12 puts the security model first, per use case. M0 has one use case, tracking and imports; the use cases of later milestones are in the proposal files. The author recorded this decision with D-65 and D-67 to D-72 after the design session that answered the implementation's questions. REQ-identity-session-opaque refines it for the implementation (SHA-256 hash, default expiries); those details are the implementation's choice, not the author's.

## Source

> After authentication the core issues an opaque, revocable session token, not a JSON Web Token (JWT). It stores only the token's hash, and the session expires after a period of inactivity and after an absolute lifetime. Each session has a scope: `read` or `owner`.

OpenFief author decisions (docs/DESIGN.md, section 3.7), design-decisions:D-66#1

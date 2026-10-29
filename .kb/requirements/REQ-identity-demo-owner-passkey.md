---
title: On a public demo anyone must be able to browse read-only, and the owner must log in with a passkey for an owner session.
status: open
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-68-2
  - lane:strict
text_ref: design-decisions:D-68#2
origin:
  kind: import
  ref: design-decisions:D-68#2
  recorded_at: '2026-10-29T05:29:58.622Z'
semantic_text: On a public demo anyone must be able to browse read-only, and the owner must log in with a passkey for an owner session.
logic_claims:
  - CLAIM-505E3F59EDAC6575
semantic_clauses:
  - On a public demo anyone must be able to browse read-only, and the owner must log in with a passkey for an owner session.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
firmness: firm
required_evidence: asserted
semantic_source_hash: 54c78f2946d583b3653e167fdeca53c7595efd1e6cedf909b80a2f2a0b987082
semantic_inventory:
  - claim_key: CLAIM-505E3F59EDAC6575
    claim_text: On a public demo anyone must be able to browse read-only, and the owner must log in with a passkey for an owner session
    role: normative
    status: modeled
    span:
      start: 0
      end: 119
id: REQ-identity-demo-owner-passkey
type: req
---
On a public demo anyone must be able to browse read-only, and the owner must log in with a passkey for an owner session.

## Context

GR-12 puts the security model first, per use case. M0 has one use case, tracking and imports; the use cases of later milestones are in the proposal files. The author recorded this decision with D-65 to D-67 and D-69 to D-72 after the design session that answered the implementation's questions. REQ-identity-demo-sessions records the first sentence of D-68, when the demo authenticator may issue read sessions; this requirement records the second.

## Source

> A demo authenticator issues `read` sessions without credentials, only when the database carries the marker of the synthetic demo dataset and the core runs with `--demo-sessions`. On a public demo anyone can browse read-only, and the owner logs in with a passkey for an `owner` session.

OpenFief author decisions (docs/DESIGN.md, section 3.7), design-decisions:D-68#2

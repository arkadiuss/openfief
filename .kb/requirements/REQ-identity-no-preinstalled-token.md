---
title: Script tokens must be created from an owner session in the UI, and no script token may be installed in advance as a Secret.
status: open
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-67-2
  - lane:strict
text_ref: design-decisions:D-67#2
origin:
  kind: import
  ref: design-decisions:D-67#2
  recorded_at: '2026-10-29T05:28:59.065Z'
semantic_text: Script tokens must be created from an owner session in the UI, and no script token may be installed in advance as a Secret.
logic_claims:
  - CLAIM-7F8EFB581645A050
semantic_clauses:
  - Script tokens must be created from an owner session in the UI, and no script token may be installed in advance as a Secret.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
firmness: firm
required_evidence: asserted
semantic_source_hash: d98711b57419bfcedc26516eb96cf4d4872dbfd514515023515c7951e754d558
semantic_inventory:
  - claim_key: CLAIM-7F8EFB581645A050
    claim_text: Script tokens must be created from an owner session in the UI, and no script token may be installed in advance as a Secret
    role: normative
    status: modeled
    span:
      start: 0
      end: 122
id: REQ-identity-no-preinstalled-token
type: req
---
Script tokens must be created from an owner session in the UI, and no script token may be installed in advance as a Secret.

## Context

GR-12 puts the security model first, per use case. M0 has one use case, tracking and imports; the use cases of later milestones are in the proposal files. The author recorded this decision with D-65, D-66 and D-68 to D-72 after the design session that answered the implementation's questions. REQ-identity-script-token-from-session records the first half of D-67, that tokens come only from an owner session with core token issue as the administrator's break-glass; this requirement records the rest.

## Source

> Script tokens (D-27) are created only from an owner session in the UI, and no token is installed in advance as a Secret. `core token issue`, run in the core's pod, stays as a break-glass for an administrator.

OpenFief author decisions (docs/DESIGN.md, section 3.7), design-decisions:D-67#2

---
title: Taking statement files must be a capability the owner sees, not a permission the grant approves.
status: open
text_ref: design-decisions:D-139#2
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-139-2
  - lane:strict
semantic_text: Taking statement files must be a capability the owner sees, not a permission the grant approves.
logic_claims:
  - CLAIM-AC8A6F8CBA43742B
semantic_clauses:
  - Taking statement files must be a capability the owner sees, not a permission the grant approves.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: f4fc4d021b8215718d4c808219cd6745f54d46e90c6fb601989907d4c90c7563
semantic_inventory:
  - claim_key: CLAIM-AC8A6F8CBA43742B
    claim_text: Taking statement files must be a capability the owner sees, not a permission the grant approves
    role: normative
    status: modeled
    span:
      start: 0
      end: 95
firmness: firm
required_evidence: asserted
origin:
  kind: import
  ref: design-decisions:D-139#2
  recorded_at: '2026-11-01T12:23:34.930Z'
id: REQ-plugins-statement-files-capability-not-permission
type: req
---
Taking statement files must be a capability the owner sees, not a permission the grant approves.

## Context

The author recorded it on 2026-11-01, answering the implementation's question #23. Reason not stated. Each instance has its own grant (D-48), which the owner approves; this capability is shown to the owner without being part of it. No end-to-end test proves it yet, so its evidence floor is asserted.

## Source

> A connector that takes statement files says so in its signed manifest, as one sending cross-check prices does (D-112). It is a capability the owner sees, not a permission the grant approves.

OpenFief author decisions (docs/DESIGN.md, section 3.6), design-decisions:D-139#2

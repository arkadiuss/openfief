---
title: The core's authentication service must never be offered on the public read-only API.
status: open
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-69-3
  - lane:strict
text_ref: design-decisions:D-69#3
origin:
  kind: import
  ref: design-decisions:D-69#3
  recorded_at: '2026-10-29T05:31:46.752Z'
semantic_text: The core's authentication service must never be offered on the public read-only API.
logic_claims:
  - CLAIM-562D2FFD592D6322
semantic_clauses:
  - The core's authentication service must never be offered on the public read-only API.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
firmness: firm
required_evidence: asserted
semantic_source_hash: f96a35e5cc59da00a34af407e3122af045c8289d124b0864c1ddfa89ff70d63e
semantic_inventory:
  - claim_key: CLAIM-562D2FFD592D6322
    claim_text: The core's authentication service must never be offered on the public read-only API
    role: normative
    status: modeled
    span:
      start: 0
      end: 83
id: REQ-identity-auth-not-public
type: req
---
The core's authentication service must never be offered on the public read-only API.

## Context

GR-12 puts the security model first, per use case. M0 has one use case, tracking and imports; the use cases of later milestones are in the proposal files. The author recorded this decision with D-65 to D-68 and D-70 to D-72 after the design session that answered the implementation's questions. D-6 keeps the public API read-only; REQ-identity-auth-callers records that the service serves view plugins only, and REQ-identity-auth-contract refines the service for the implementation.

## Source

> The built-in UI is a backend-for-frontend: the browser talks only to the UI view plugin, and the plugin calls the core's authentication service (`openfief.identity.v1`) for login and sessions. That service is offered to view plugins only, never on the public read-only API (D-6).

OpenFief author decisions (docs/DESIGN.md, section 3.7), design-decisions:D-69#3

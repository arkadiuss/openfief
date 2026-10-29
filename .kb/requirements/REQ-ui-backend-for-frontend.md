---
title: 'The built-in UI must be a backend-for-frontend: the browser must talk only to the UI view plugin, and the plugin must call the core''s authentication service, openfief.identity.v1, for login and sessions.'
status: open
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-69-1
  - lane:strict
text_ref: design-decisions:D-69#1
origin:
  kind: import
  ref: design-decisions:D-69#1
  recorded_at: '2026-10-29T05:30:45.548Z'
semantic_text: 'The built-in UI must be a backend-for-frontend: the browser must talk only to the UI view plugin, and the plugin must call the core''s authentication service, openfief.identity.v1, for login and sessions.'
logic_claims:
  - CLAIM-6FE97CA58DA32821
semantic_clauses:
  - 'The built-in UI must be a backend-for-frontend: the browser must talk only to the UI view plugin, and the plugin must call the core''s authentication service, openfief.identity.v1, for login and sessions.'
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
firmness: firm
required_evidence: asserted
semantic_source_hash: a5db13e513cd595295be73bb62cebe06b2bed53fd1a3296afc01604f72f1a76e
semantic_inventory:
  - claim_key: CLAIM-6FE97CA58DA32821
    claim_text: 'The built-in UI must be a backend-for-frontend: the browser must talk only to the UI view plugin, and the plugin must call the core''s authentication service, openfief.identity.v1, for login and sessions'
    role: normative
    status: modeled
    span:
      start: 0
      end: 202
id: REQ-ui-backend-for-frontend
type: req
---
The built-in UI must be a backend-for-frontend: the browser must talk only to the UI view plugin, and the plugin must call the core's authentication service, openfief.identity.v1, for login and sessions.

## Context

GR-12 puts the security model first, per use case. M0 has one use case, tracking and imports; the use cases of later milestones are in the proposal files. The author recorded this decision with D-65 to D-68 and D-70 to D-72 after the design session that answered the implementation's questions. The UI's existing requirements on signing in through the AuthService, its session cookie and its ServiceAccount token refine this one for the implementation.

## Source

> The built-in UI is a backend-for-frontend: the browser talks only to the UI view plugin, and the plugin calls the core's authentication service (`openfief.identity.v1`) for login and sessions. That service is offered to view plugins only, never on the public read-only API (D-6).

OpenFief author decisions (docs/DESIGN.md, section 3.7), design-decisions:D-69#1

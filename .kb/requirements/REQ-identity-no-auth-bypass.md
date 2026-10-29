---
title: Every build of OpenFief, development builds included, must have no fixed token and no other way around authentication.
status: open
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-84-1
  - lane:strict
text_ref: design-decisions:D-84#1
origin:
  kind: import
  ref: design-decisions:D-84#1
  recorded_at: '2026-10-29T08:21:32.861Z'
semantic_text: Every build of OpenFief, development builds included, must have no fixed token and no other way around authentication.
logic_claims:
  - CLAIM-8DDF24D2356BCDDD
semantic_clauses:
  - Every build of OpenFief, development builds included, must have no fixed token and no other way around authentication.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
firmness: firm
required_evidence: asserted
semantic_source_hash: e5b5596070217f6779053d7079edbe51b9d991df7937e76fb78ebf959d84acaa
semantic_inventory:
  - claim_key: CLAIM-8DDF24D2356BCDDD
    claim_text: Every build of OpenFief, development builds included, must have no fixed token and no other way around authentication
    role: normative
    status: modeled
    span:
      start: 0
      end: 117
id: REQ-identity-no-auth-bypass
type: req
---
Every build of OpenFief, development builds included, must have no fixed token and no other way around authentication.

## Context

D-65 to D-69 make the owner authenticate with a passkey and then use an opaque session, and D-27 has plugins and scripts present their own tokens. D-84 rules out every shortcut around that, in every build and even for development: no fixed token and no other bypass. Local development of the UI runs against a local cluster instead (REQ-ui-local-development-cluster). The author recorded D-83 to D-85 after the design session that answered the implementation's questions.

## Source

> No build of OpenFief has a fixed token or any other way around authentication, not even for development. Local development of the UI runs against a local cluster such as kind or k3d.

OpenFief author decisions (docs/DESIGN.md, section 3.7), design-decisions:D-84#1

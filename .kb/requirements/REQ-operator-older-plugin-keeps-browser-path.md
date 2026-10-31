---
title: When two plugins claim a browser path at once, the older Plugin, by creation time and then by name, must keep it.
status: open
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-125-1
  - lane:strict
text_ref: design-decisions:D-125#1
origin:
  kind: import
  ref: design-decisions:D-125#1
  recorded_at: '2026-10-31T14:34:38.000Z'
semantic_text: When two plugins claim a browser path at once, the older Plugin, by creation time and then by name, must keep it.
logic_claims:
  - CLAIM-EA09F68FCC9B4ECB
semantic_clauses:
  - When two plugins claim a browser path at once, the older Plugin, by creation time and then by name, must keep it.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
firmness: firm
required_evidence: asserted
semantic_source_hash: 2a00e5f0280b503094cc95221fc05b9f307790a781d71250ddf313dace1226e9
semantic_inventory:
  - claim_key: CLAIM-EA09F68FCC9B4ECB
    claim_text: When two plugins claim a browser path at once, the older Plugin, by creation time and then by name, must keep it
    role: condition
    status: modeled
    span:
      start: 0
      end: 112
id: REQ-operator-older-plugin-keeps-browser-path
type: req
---
When two plugins claim a browser path at once, the older Plugin, by creation time and then by name, must keep it.

## Context

D-114 lets at most one plugin instance hold a browser path in M0 (REQ-plugins-one-browser-view), and the operator doesn't run a second one (REQ-operator-refuses-second-browser-view). Two such plugins applied at once, as GitOps may, can both start before either's status records a path, and then each sees the other holding one. The implementation raised this as question #9 on docs/implementation-questions.md and lets the older Plugin keep the path; on 2026-10-31 the author confirmed that as D-125, rather than letting the newer one keep it or refusing both. REQ-operator-contested-browser-path-stops says what happens to the other.

## Source

> When two plugins claim a browser path at once, the older Plugin, by creation time and then by name, keeps it; the other stops, with the reason in its status.

OpenFief author decisions (docs/DESIGN.md, section 3.7), design-decisions:D-125#1

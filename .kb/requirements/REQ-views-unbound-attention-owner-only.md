---
title: Once viewers exist, only the owner may see an attention item about something tied to no account.
status: open
text_ref: design-decisions:D-165#2
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-165-2
  - lane:strict
semantic_text: Once viewers exist, only the owner may see an attention item about something tied to no account.
logic_claims:
  - CLAIM-22E4727247D56F53
semantic_clauses:
  - Once viewers exist, only the owner may see an attention item about something tied to no account.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 3bfa6a5eab9db90d740db49e4831996052273529db3dc794abaf5dfbde961299
semantic_inventory:
  - claim_key: CLAIM-22E4727247D56F53
    claim_text: Once viewers exist, only the owner may see an attention item about something tied to no account
    role: normative
    status: modeled
    span:
      start: 0
      end: 95
firmness: firm
required_evidence: asserted
origin:
  kind: import
  ref: design-decisions:D-165#2
  recorded_at: '2026-11-02T00:55:40.810Z'
id: REQ-views-unbound-attention-owner-only
type: req
---
Once viewers exist, only the owner may see an attention item about something tied to no account.

## Context

The author recorded it on 2026-11-01, when planning viewers for the second alpha. Reason not stated. Viewers come with the second alpha (D-65) and see only the accounts granted to them (D-163). No end-to-end test proves it yet, so its evidence floor is asserted.

## Source

> An attention item about something tied to no account, such as an unmapped broker account, names only its connector instance on the read API. Once viewers exist (D-65), only the owner sees such items.

OpenFief author decisions (docs/DESIGN.md, section 3.7), design-decisions:D-165#2

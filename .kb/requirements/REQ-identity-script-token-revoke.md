---
title: Revoking a script token must refuse its secret from then on, keep the time of the first revocation when revoked again, fail for an unknown or malformed ID, and listing must show every token, expired and revoked ones included, with its status.
status: open
text_ref: docs/DESIGN.md#D-27
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-27
  - lane:strict
semantic_text: Revoking a script token must refuse its secret from then on, keep the time of the first revocation when revoked again, fail for an unknown or malformed ID, and listing must show every token, expired and revoked ones included, with its status.
logic_claims:
  - CLAIM-4BC942053859F808
semantic_clauses:
  - Revoking a script token must refuse its secret from then on, keep the time of the first revocation when revoked again, fail for an unknown or malformed ID, and listing must show every token, expired and revoked ones included, with its status.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 70185530a1457ca8fd3fd6cd012a4780824eaf68f150334e2a8e389bb5c541eb
semantic_inventory:
  - claim_key: CLAIM-4BC942053859F808
    claim_text: Revoking a script token must refuse its secret from then on, keep the time of the first revocation when revoked again, fail for an unknown or malformed ID, and listing must show every token, expired and revoked ones included, with its status
    role: normative
    status: modeled
    span:
      start: 0
      end: 241
origin:
  kind: agent
  recorded_at: '2026-10-28T14:11:31.206Z'
id: REQ-identity-script-token-revoke
type: req
---
Revoking a script token must refuse its secret from then on, keep the time of the first revocation when revoked again, fail for an unknown or malformed ID, and listing must show every token, expired and revoked ones included, with its status.

## Context

A personal token that leaks must be stoppable at once; keeping revoked and expired tokens listed lets the owner see what existed. Reason for keeping the first time: an audit trail should not move. Modeled by Claude Code for the core task: the author asked for minimal script tokens so the read API can run on the demo; the author has not reviewed this requirement yet.

## Source

> Scripts get separate, scoped, expiring, read-only personal tokens.

docs/DESIGN.md D-27

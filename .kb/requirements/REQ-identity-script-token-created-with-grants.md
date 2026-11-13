---
title: 'A script token is created with the grants, the lifetime and the MCP flag the owner chose: the token, its grants and their audit entries are stored in one transaction, and then the grants are written to SpiceDB; if that write fails the token is revoked and audited and no secret is shown.'
status: open
origin:
  kind: agent
  ref: 'S5 task brief: the Access page, script tokens and view plugins (D-209, D-210), 2026-11-13'
  recorded_at: '2026-11-13T18:00:00.000Z'
firmness: default
required_evidence: unverified
semantic_text: 'A script token is created with the grants, the lifetime and the MCP flag the owner chose: the token, its grants and their audit entries are stored in one transaction, and then the grants are written to SpiceDB; if that write fails the token is revoked and audited and no secret is shown.'
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 300016378e28971ed9caeae67682cbfd2bb8bbeb75be84871a5c2cf6335ba8ea
semantic_inventory:
  - claim_key: CLAIM-41AFF53F2F6513F1
    claim_text: 'A script token is created with the grants, the lifetime and the MCP flag the owner chose: the token, its grants and their audit entries are stored in one transaction, and then the grants are written to SpiceDB'
    role: descriptive
    status: missing
    span:
      start: 0
      end: 209
  - claim_key: CLAIM-386770C2C5C4337E
    claim_text: if that write fails the token is revoked and audited and no secret is shown
    role: condition
    status: ontology_gap
    span:
      start: 211
      end: 286
logic_claims:
  - CLAIM-41AFF53F2F6513F1
  - CLAIM-386770C2C5C4337E
id: REQ-identity-script-token-created-with-grants
type: req
---
A script token is created with the grants, the lifetime and the MCP flag the owner chose: the token, its grants and their audit entries are stored in one transaction, and then the grants are written to SpiceDB; if that write fails the token is revoked and audited and no secret is shown.

## Context

The order is the one core token issue already follows. A token that is to hold grants asks SpiceDB whether it answers before anything is stored; a target that does not exist refuses the token before anything is stored.
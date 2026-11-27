---
title: The read API must authenticate a bearer token as its holder when it is a valid session of any scope (the owner, a viewer or the demo) or a valid read-scoped script token, and must refuse a session it can't place as one of these.
status: open
firmness: firm
required_evidence: asserted
text_ref: design-decisions:D-83#1, design-decisions:D-65
semantic_text: The read API must authenticate a bearer token as its holder when it is a valid session of any scope (the owner, a viewer or the demo) or a valid read-scoped script token, and must refuse a session it can't place as one of these.
origin:
  kind: agent
  ref: the coordinator relayed on 27 Nov 2026 that the author approved this reword directly (viewers authenticate as viewers, D-65; the S1 fix made unknown sessions fail closed); approved_by is for the author to record
  recorded_at: '2026-11-27T10:01:31.517Z'
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 4ccedd3a59917d582117066a5382f2148751bfd2ce0f412756f7d469ebc56df7
semantic_inventory:
  - claim_key: CLAIM-1DC99BA9E2E2BB6B
    claim_text: The read API must authenticate a bearer token as its holder when it is a valid session of any scope (the owner, a viewer or the demo) or a valid read-scoped script token
    role: normative
    status: missing
    span:
      start: 0
      end: 169
    subjects:
      - identity.read_api_session_holder
  - claim_key: CLAIM-F3CE761FD52A6844
    claim_text: must refuse a session it can't place as one of these
    role: normative
    status: missing
    span:
      start: 175
      end: 227
    subjects:
      - identity.read_api_session_holder
logic_claims:
  - CLAIM-1DC99BA9E2E2BB6B
  - CLAIM-F3CE761FD52A6844
id: REQ-identity-read-api-sessions-v2
type: req
---
The read API must authenticate a bearer token as its holder when it is a valid session of any scope (the owner, a viewer or the demo) or a valid read-scoped script token, and must refuse a session it can't place as one of these.

## Why this changed

The author approved this rewording directly on 27 Nov 2026 (relayed by the coordinator). REQ-identity-read-api-sessions said a valid session of any scope authenticates "as the owner", but viewers authenticate as viewers (D-65), and the fix of S1 (a viewer's session resolving as the owner) made a session the core can't place as the owner's, a viewer's or the demo's fail closed. The old requirement is superseded, and its code and scenario links moved here.

## Context

D-83 settles which callers the read API accepts: sessions of either scope, script tokens and view plugins' own tokens (see REQ-identity-read-api-view-plugin-tokens and REQ-identity-read-api-refuses-connector-tokens). identityapp.PrincipalOfSession builds who a session stands for, and Principal.Caller is the owner only for the owner's own session.

## Source

> The read API accepts sessions of either scope, script tokens, and view plugins' own tokens. It refuses connector tokens: the core calls connectors (D-49), and a connector never needs to read the owner's record.

OpenFief author decisions (docs/DESIGN.md, section 3.7), design-decisions:D-83#1, and D-65.

---
title: The Access page lists the people, the owner and each viewer with their grants and sessions, who are invited by a one-time link to enroll their own passkey and revoked from there.
status: open
origin:
  kind: import
  ref: design-decisions:D-210#2
  recorded_at: '2026-11-13T22:00:00.000Z'
id: REQ-ui-access-people
type: req
firmness: firm
semantic_text: The Access page lists the people, the owner and each viewer with their grants and sessions, who are invited by a one-time link to enroll their own passkey and revoked from there.
text_ref: design-decisions:D-210#2
required_evidence: asserted
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 07a9c0e8af5bbd4af13de095c13939e49e5aa0a4e5e4c40c3dc17922ceccd2a7
semantic_inventory:
  - claim_key: CLAIM-90DFAE47392568ED
    claim_text: The Access page lists the people, the owner and each viewer with their grants and sessions, who are invited by a one-time link to enroll their own passkey and revoked from there
    role: descriptive
    status: missing
    span:
      start: 0
      end: 177
logic_claims:
  - CLAIM-90DFAE47392568ED
---
The Access page lists the people, the owner and each viewer with their grants and sessions, who are invited by a one-time link to enroll their own passkey and revoked from there.

## Context

D-210 (the author's, 2026-11-12). The owner-only People section of the Access page: the owner row, each viewer with name, status, grants in words, active sessions and passkeys, editing grants through the shared grants form, an invitation form whose one-time link is shown once with its expiry, and Revoke after a confirmation; revoked viewers stay listed.

## Source

> Settings has an Access page: the people (the owner, and each viewer with their grants and sessions, invited by a one-time link to enroll their own passkey and revoked from there), the script tokens (each with its grants, its expiry and whether it may use MCP), the view plugins with their grants, and the history of every grant and revocation.

OpenFief author decisions (docs/DESIGN.md), design-decisions:D-210#1
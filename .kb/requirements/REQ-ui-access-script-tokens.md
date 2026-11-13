---
title: The Access page lists the script tokens, each with its grants, its expiry and whether it may use MCP.
status: open
text_ref: design-decisions:D-210#2
origin:
  kind: import
  ref: design-decisions:D-210#2
  recorded_at: '2026-11-13T18:00:00.000Z'
firmness: firm
required_evidence: asserted
semantic_text: The Access page lists the script tokens, each with its grants, its expiry and whether it may use MCP.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 2fb8ddb8a1c7aa25ba97693ccab3e3afcac24a5b828baa647198ac4772aa5a81
semantic_inventory:
  - claim_key: CLAIM-88D63305F54010E7
    claim_text: The Access page lists the script tokens, each with its grants, its expiry and whether it may use MCP
    role: descriptive
    status: missing
    span:
      start: 0
      end: 100
logic_claims:
  - CLAIM-88D63305F54010E7
id: REQ-ui-access-script-tokens
type: req
---
The Access page lists the script tokens, each with its grants, its expiry and whether it may use MCP.

## Context

D-210. The owner creates a token there with a name, a lifetime, the MCP flag (stored only: D-213 is not built) and the grants it starts with, changes its grants later, and revokes it. Its secret is shown once, when it is made.

## Source

> Settings has an Access page: the people (the owner, and each viewer with their grants and sessions, invited by a one-time link to enroll their own passkey and revoked from there), the script tokens (each with its grants, its expiry and whether it may use MCP), the view plugins with their grants, and the history of every grant and revocation.

OpenFief author decisions (docs/DESIGN.md), design-decisions:D-210#2
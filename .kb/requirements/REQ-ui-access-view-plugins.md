---
title: The Access page lists the view plugins with their grants.
status: open
text_ref: design-decisions:D-210#4
origin:
  kind: import
  ref: design-decisions:D-210#4
  recorded_at: '2026-11-13T18:00:00.000Z'
firmness: firm
required_evidence: asserted
semantic_text: The Access page lists the view plugins with their grants.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 47514436667941f3dcb499c0647ca66663c90c42dbc25bcaa7b50fa5097b34a9
semantic_inventory:
  - claim_key: CLAIM-12A217F20120214F
    claim_text: The Access page lists the view plugins with their grants
    role: descriptive
    status: missing
    span:
      start: 0
      end: 56
logic_claims:
  - CLAIM-12A217F20120214F
id: REQ-ui-access-view-plugins
type: req
---
The Access page lists the view plugins with their grants.

## Context

D-210. The built-in UI is left out: it reads with the signed-in person's session, never its own identity. The owner adds and removes a plugin's grants there, and the Inbox item of a plugin that can read nothing links to this section.

## Source

> Settings has an Access page: the people (the owner, and each viewer with their grants and sessions, invited by a one-time link to enroll their own passkey and revoked from there), the script tokens (each with its grants, its expiry and whether it may use MCP), the view plugins with their grants, and the history of every grant and revocation.

OpenFief author decisions (docs/DESIGN.md), design-decisions:D-210#4
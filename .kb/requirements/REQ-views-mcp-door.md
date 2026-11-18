---
title: 'The core answers MCP at /mcp on its browser port only while the one switch and the MCP toggle are both on, checked at every request, and only to a script token the owner allowed to use MCP: no token or an invalid, expired or revoked one is 401 with a Bearer challenge, a valid token without the flag 403, and with either off 404, none of them with a tool list.'
status: open
firmness: default
required_evidence: unverified
origin:
  kind: agent
  ref: 'D-213 with D-212, D-211, D-209 and D-203: the MCP server brief from the design session, 2026-11-18'
  recorded_at: '2026-11-18T09:00:00.000Z'
semantic_text: 'The core answers MCP at /mcp on its browser port only while the one switch and the MCP toggle are both on, checked at every request, and only to a script token the owner allowed to use MCP: no token or an invalid, expired or revoked one is 401 with a Bearer challenge, a valid token without the flag 403, and with either off 404, none of them with a tool list.'
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: cab108ed6b33f9e021ed24280edb1c345e32c80b027a9781cc86e1a90b1f20fc
semantic_inventory:
  - claim_key: CLAIM-F60194DAEEC7CF9F
    claim_text: 'The core answers MCP at /mcp on its browser port only while the one switch and the MCP toggle are both on, checked at every request, and only to a script token the owner allowed to use MCP: no token or an invalid, expired or revoked one is 401 with a Bearer challenge, a valid token without the flag 403, and with either off 404, none of them with a tool list'
    role: descriptive
    status: missing
    span:
      start: 0
      end: 359
logic_claims:
  - CLAIM-F60194DAEEC7CF9F
semantic_clauses:
  - 'The core answers MCP at /mcp on its browser port only while the one switch and the MCP toggle are both on, checked at every request, and only to a script token the owner allowed to use MCP: no token or an invalid, expired or revoked one is 401 with a Bearer challenge, a valid token without the flag 403, and with either off 404, none of them with a tool list'
id: REQ-views-mcp-door
type: req
---
The core answers MCP at /mcp on its browser port only while the one switch and the MCP toggle are both on, checked at every request, and only to a script token the owner allowed to use MCP: no token or an invalid, expired or revoked one is 401 with a Bearer challenge, a valid token without the flag 403, and with either off 404, none of them with a tool list.

## Context

D-213 lets the core serve MCP for external agents behind a script token the owner allowed to use MCP, and D-212 says that with the switch off no MCP endpoint answers. The switch and the toggle are read at every request, so turning either off stops the answers at once. The endpoint is the core's own path (D-81), so no view plugin can claim it.

## Source

> the core can serve MCP at /mcp on its browser port ... A client authenticates with a script token the owner allowed to use MCP ... With the switch off, nothing calls a model and no MCP endpoint answers.

Source: OpenFief author decisions (docs/DESIGN.md), D-213 with D-212, D-211, D-209 and D-203: the MCP server brief from the design session, 2026-11-18

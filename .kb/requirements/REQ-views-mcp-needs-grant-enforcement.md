---
title: The MCP server answers 503 while the switch and the MCP toggle are on and the authorizer does not enforce script token grants, and 404 while either is off whatever the authorizer, so it is never served where grants are ignored.
status: open
firmness: default
required_evidence: unverified
origin:
  kind: agent
  ref: the design session's answers on D-213, D-209, D-165 and D-199, relayed by the coordinator, 2026-11-18
  recorded_at: '2026-11-18T17:00:00.000Z'
semantic_text: The MCP server answers 503 while the switch and the MCP toggle are on and the authorizer does not enforce script token grants, and 404 while either is off whatever the authorizer, so it is never served where grants are ignored.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: c3a5bf0f69bffca79a567f1a0b062f2087d2d65354688cecd86d8e4960a25cec
semantic_inventory:
  - claim_key: CLAIM-E0DA40938D9B557F
    claim_text: The MCP server answers 503 while the switch and the MCP toggle are on and the authorizer does not enforce script token grants, and 404 while either is off whatever the authorizer, so it is never served where grants are ignored
    role: descriptive
    status: missing
    span:
      start: 0
      end: 226
semantic_clauses:
  - The MCP server answers 503 while the switch and the MCP toggle are on and the authorizer does not enforce script token grants, and 404 while either is off whatever the authorizer, so it is never served where grants are ignored
logic_claims:
  - CLAIM-E0DA40938D9B557F
id: REQ-views-mcp-needs-grant-enforcement
type: req
---
The MCP server answers 503 while the switch and the MCP toggle are on and the authorizer does not enforce script token grants, and 404 while either is off whatever the authorizer, so it is never served where grants are ignored.

## Context

The authorizers say EnforcesTokenGrants: GrantPolicy and SpiceDB do, the bare Policy does not.

## Source

> whose grants (D-209) limit what every tool sees.

Source: the design session's answers on D-213, D-209, D-165 and D-199, relayed by the coordinator, 2026-11-18

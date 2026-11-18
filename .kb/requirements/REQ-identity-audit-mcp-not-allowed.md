---
title: A valid script token the owner did not allow to use MCP is refused with 403 and an audit entry of the action mcp_call with the outcome refused, the token as actor and no tool named, while a request refused with 401 writes no entry and is counted, and one refused with 404 or 503 writes nothing.
status: open
firmness: default
required_evidence: unverified
origin:
  kind: agent
  ref: the design session's answers on D-213, D-209, D-165 and D-199, relayed by the coordinator, 2026-11-18
  recorded_at: '2026-11-18T17:00:00.000Z'
semantic_text: A valid script token the owner did not allow to use MCP is refused with 403 and an audit entry of the action mcp_call with the outcome refused, the token as actor and no tool named, while a request refused with 401 writes no entry and is counted, and one refused with 404 or 503 writes nothing.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 740d71d8dabef11040a8d5bb3b65dd40a04fed5a49e0dc12813d9c7bcbd63687
semantic_inventory:
  - claim_key: CLAIM-FF5E1A6EA42F763D
    claim_text: A valid script token the owner did not allow to use MCP is refused with 403 and an audit entry of the action mcp_call with the outcome refused, the token as actor and no tool named, while a request refused with 401 writes no entry and is counted, and one refused with 404 or 503 writes nothing
    role: descriptive
    status: missing
    span:
      start: 0
      end: 293
semantic_clauses:
  - A valid script token the owner did not allow to use MCP is refused with 403 and an audit entry of the action mcp_call with the outcome refused, the token as actor and no tool named, while a request refused with 401 writes no entry and is counted, and one refused with 404 or 503 writes nothing
logic_claims:
  - CLAIM-FF5E1A6EA42F763D
id: REQ-identity-audit-mcp-not-allowed
type: req
---
A valid script token the owner did not allow to use MCP is refused with 403 and an audit entry of the action mcp_call with the outcome refused, the token as actor and no tool named, while a request refused with 401 writes no entry and is counted, and one refused with 404 or 503 writes nothing.

## Context

D-211 records access events; a 401 can be sent by anyone, so it is counted by the metric openfief.mcp.unauthenticated instead.

## Source

> each call enters the audit log as usage only (D-211).

Source: the design session's answers on D-213, D-209, D-165 and D-199, relayed by the coordinator, 2026-11-18

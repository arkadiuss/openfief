---
title: The MCP server offers exactly the read-only tools the Ask box's model is offered, runs each through the same answerer and the read API with the calling token's own identity so that token's grants limit what every tool sees, and changes nothing.
status: open
firmness: default
required_evidence: unverified
origin:
  kind: agent
  ref: 'D-213 with D-212, D-211, D-209 and D-203: the MCP server brief from the design session, 2026-11-18'
  recorded_at: '2026-11-18T09:00:00.000Z'
semantic_text: The MCP server offers exactly the read-only tools the Ask box's model is offered, runs each through the same answerer and the read API with the calling token's own identity so that token's grants limit what every tool sees, and changes nothing.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 243b92f5976713d8983e827782fa66fd1491e0c2ad4837635ed1e878020b57be
semantic_inventory:
  - claim_key: CLAIM-C0B798F82D40E678
    claim_text: The MCP server offers exactly the read-only tools the Ask box's model is offered, runs each through the same answerer and the read API with the calling token's own identity so that token's grants limit what every tool sees, and changes nothing
    role: rationale
    status: nonlogical
    span:
      start: 0
      end: 243
semantic_clauses:
  - The MCP server offers exactly the read-only tools the Ask box's model is offered, runs each through the same answerer and the read API with the calling token's own identity so that token's grants limit what every tool sees, and changes nothing
id: REQ-views-mcp-tools-are-ask-box-tools
type: req
---
The MCP server offers exactly the read-only tools the Ask box's model is offered, runs each through the same answerer and the read API with the calling token's own identity so that token's grants limit what every tool sees, and changes nothing.

## Context

D-213 offers the same read-only answerers as tools, and D-209 limits a script token by its grants. The tools are the Ask box's one table (D-203), called with the caller's token as bearer, never the core's own identity. Under the default policy authorizer a script token's grants are not enforced (the read API's own behavior, D-164); they limit what a tool sees once SpiceDB answers.

## Source

> offering the same read-only answerers as tools ... whose grants (D-209) limit what every tool sees. It never changes anything.

Source: OpenFief author decisions (docs/DESIGN.md), D-213 with D-212, D-211, D-209 and D-203: the MCP server brief from the design session, 2026-11-18

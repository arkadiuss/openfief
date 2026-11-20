---
title: Each MCP tool call enters the audit log as one entry whose actor is the script token and that names the tool and how the call ended, never an argument, a result, an amount or a question's text, and a call whose entry cannot be written shows nothing.
status: open
firmness: default
required_evidence: unverified
origin:
  kind: agent
  ref: 'D-213 with D-212, D-211, D-209 and D-203: the MCP server brief from the design session, 2026-11-18'
  recorded_at: '2026-11-18T09:00:00.000Z'
semantic_text: Each MCP tool call enters the audit log as one entry whose actor is the script token and that names the tool and how the call ended, never an argument, a result, an amount or a question's text, and a call whose entry cannot be written shows nothing.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 1fa5c199283a2253485a0a3ce0d6807b6f91790939300c8bbd8c5e8f7851712c
semantic_inventory:
  - claim_key: CLAIM-1F61AC753CBFF8B0
    claim_text: Each MCP tool call enters the audit log as one entry whose actor is the script token and that names the tool and how the call ended, never an argument, a result, an amount or a question's text, and a call whose entry cannot be written shows nothing
    role: normative
    status: ontology_gap
    span:
      start: 0
      end: 248
    subjects:
      - audit.mcp_call_entries
logic_claims:
  - CLAIM-1F61AC753CBFF8B0
semantic_clauses:
  - Each MCP tool call enters the audit log as one entry whose actor is the script token and that names the tool and how the call ended, never an argument, a result, an amount or a question's text, and a call whose entry cannot be written shows nothing
id: REQ-identity-audit-mcp-calls
type: req
---
Each MCP tool call enters the audit log as one entry whose actor is the script token and that names the tool and how the call ended, never an argument, a result, an amount or a question's text, and a call whose entry cannot be written shows nothing.

## Context

D-211 records model and MCP calls as usage, naming who, what and the outcome and never an amount or a text (GR-12). A script token is the actor of its MCP calls alone, since it still changes nothing; the fail-closed rule is the model calls' (REQ-models-calls-audited-fail-closed).

## Source

> each call enters the audit log as usage only (D-211).

Source: OpenFief author decisions (docs/DESIGN.md), D-213 with D-212, D-211, D-209 and D-203: the MCP server brief from the design session, 2026-11-18

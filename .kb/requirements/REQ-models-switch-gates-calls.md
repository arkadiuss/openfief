---
title: With the switch off, nothing calls a model and no MCP endpoint answers.
status: open
text_ref: design-decisions:D-212#2
firmness: firm
origin:
  kind: import
  ref: design-decisions:D-212#2
  recorded_at: '2026-11-13T11:28:47.077Z'
semantic_text: With the switch off, nothing calls a model and no MCP endpoint answers.
semantic_source_field: semantic_text
semantic_source_hash: 4d4b4cfbce256a7f24232f6e1fbae99d0eb760bac147ace4d1f852b2ff4749e4
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_inventory:
  - claim_key: CLAIM-4918566C75EEC665
    claim_text: With the switch off, nothing calls a model and no MCP endpoint answers
    role: descriptive
    status: missing
    span:
      start: 0
      end: 70
logic_claims:
  - CLAIM-4918566C75EEC665
semantic_clauses:
  - With the switch off, nothing calls a model and no MCP endpoint answers
required_evidence: asserted
id: REQ-models-switch-gates-calls
type: req
---
With the switch off, nothing calls a model and no MCP endpoint answers.

## Context

The model router enforces it before it reads a request or reaches any provider, and a feature whose own toggle is off is refused the same way. The MCP half is not built yet (D-213).

## Source

> With the switch off, nothing calls a model and no MCP endpoint answers.

Source: OpenFief author decisions (DESIGN.md section 10) - D-212#2
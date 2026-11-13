---
title: Settings has an Access page.
status: open
text_ref: design-decisions:D-210#1
origin:
  kind: import
  ref: design-decisions:D-210#1
  recorded_at: '2026-11-13T18:00:00.000Z'
firmness: firm
required_evidence: asserted
semantic_text: Settings has an Access page.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: f4bb4389d3dd4654d9bd4a549fecc7bc7bf5941442b1e5ef77f3a7e12bbb454c
semantic_inventory:
  - claim_key: CLAIM-7D4F5B2B144F1EB0
    claim_text: Settings has an Access page
    role: descriptive
    status: missing
    span:
      start: 0
      end: 27
logic_claims:
  - CLAIM-7D4F5B2B144F1EB0
id: REQ-ui-access-page
type: req
---
Settings has an Access page.

## Context

The author decided it on 2026-11-12 (D-210). The built-in UI serves it at /access, in the owner's navigation only. The people and the history on it come with the viewers and the audit view; the script tokens and the view plugins are here.

## Source

> Settings has an Access page: the people (the owner, and each viewer with their grants and sessions, invited by a one-time link to enroll their own passkey and revoked from there), the script tokens (each with its grants, its expiry and whether it may use MCP), the view plugins with their grants, and the history of every grant and revocation.

OpenFief author decisions (docs/DESIGN.md), design-decisions:D-210#1
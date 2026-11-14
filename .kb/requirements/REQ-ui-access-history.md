---
title: The Access page shows the history of every grant and revocation.
status: open
origin:
  kind: import
  ref: design-decisions:D-210#5
  recorded_at: '2026-11-14T00:30:00.000Z'
firmness: firm
semantic_text: The Access page shows the history of every grant and revocation.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: cc72fbe8abc8abafa4064dfcd6fab6b6e46876533d49f56b6b646f2d2050efc4
semantic_inventory:
  - claim_key: CLAIM-C45D13D3AC3A6621
    claim_text: The Access page shows the history of every grant and revocation
    role: descriptive
    status: missing
    span:
      start: 0
      end: 63
logic_claims:
  - CLAIM-C45D13D3AC3A6621
text_ref: design-decisions:D-210#5
id: REQ-ui-access-history
type: req
required_evidence: asserted
---
The Access page shows the history of every grant and revocation.

## Context

D-210 (the author's, 2026-11-12), fifth clause. The owner-only History section lists the grants and revocations of viewers, script tokens and view plugins, with the links that invite viewers, newest first and paged, from the audit log (D-211).

## Source

> Settings has an Access page: the people (...), the script tokens (...), the view plugins with their grants, and the history of every grant and revocation.

OpenFief author decisions (docs/DESIGN.md), design-decisions:D-210#5
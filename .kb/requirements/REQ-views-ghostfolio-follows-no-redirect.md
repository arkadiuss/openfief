---
title: The Ghostfolio plugin must never follow a redirect.
status: open
text_ref: design-decisions:D-188#2
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-188-2
  - lane:strict
semantic_text: The Ghostfolio plugin must never follow a redirect.
logic_claims:
  - CLAIM-7F3E32770A0CC496
semantic_clauses:
  - The Ghostfolio plugin must never follow a redirect.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 1f63bd08849617c7fbd964eeb44a0380ef7115411820244e6d1307b2d6400ba3
semantic_inventory:
  - claim_key: CLAIM-7F3E32770A0CC496
    claim_text: The Ghostfolio plugin must never follow a redirect
    role: normative
    status: modeled
    span:
      start: 0
      end: 50
firmness: firm
required_evidence: asserted
origin:
  kind: import
  ref: design-decisions:D-188#2
  recorded_at: '2026-11-03T14:31:50.956Z'
id: REQ-views-ghostfolio-follows-no-redirect
type: req
---
The Ghostfolio plugin must never follow a redirect.

## Context

The author decided it on 2026-11-03, in the review's second pass before the alpha tag. A 307 or 308 redirect sends the request's body again, and the login's body carries the security token. No end-to-end test proves it yet, so its evidence floor is asserted.

## Source

> The Ghostfolio plugin checks that the Ghostfolio it talks to is a version it supports before it changes or deletes anything there, and never follows a redirect.

OpenFief author decisions (docs/DESIGN.md, section 3.1), design-decisions:D-188#2

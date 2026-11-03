---
title: The Ghostfolio plugin must check that the Ghostfolio it talks to is a version it supports before it changes or deletes anything there.
status: open
text_ref: design-decisions:D-188#1
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-188-1
  - lane:strict
semantic_text: The Ghostfolio plugin must check that the Ghostfolio it talks to is a version it supports before it changes or deletes anything there.
logic_claims:
  - CLAIM-45C8546E8F31333D
semantic_clauses:
  - The Ghostfolio plugin must check that the Ghostfolio it talks to is a version it supports before it changes or deletes anything there.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: f1d1a10a9325099f8c6308afbb5f24596fd69e7f981788218dc03b7f379ac301
semantic_inventory:
  - claim_key: CLAIM-45C8546E8F31333D
    claim_text: The Ghostfolio plugin must check that the Ghostfolio it talks to is a version it supports before it changes or deletes anything there
    role: normative
    status: modeled
    span:
      start: 0
      end: 133
firmness: firm
required_evidence: asserted
origin:
  kind: import
  ref: design-decisions:D-188#1
  recorded_at: '2026-11-03T14:31:39.016Z'
id: REQ-views-ghostfolio-checks-version-before-writes
type: req
---
The Ghostfolio plugin must check that the Ghostfolio it talks to is a version it supports before it changes or deletes anything there.

## Context

The author decided it on 2026-11-03, in the review's second pass before the alpha tag. D-64 pins a range of Ghostfolio versions because Ghostfolio renames its internal endpoints within /v1; a plugin writing to another version could change or delete the wrong things. No end-to-end test proves it yet, so its evidence floor is asserted.

## Source

> The Ghostfolio plugin checks that the Ghostfolio it talks to is a version it supports before it changes or deletes anything there, and never follows a redirect.

OpenFief author decisions (docs/DESIGN.md, section 3.1), design-decisions:D-188#1

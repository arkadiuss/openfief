---
title: The Ghostfolio plugin must put the OpenFief transaction ID in each activity's comment.
status: open
text_ref: design-decisions:D-4#2
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-4-2
  - lane:strict
semantic_text: The Ghostfolio plugin must put the OpenFief transaction ID in each activity's comment.
logic_claims:
  - CLAIM-7F3B612B362716C7
semantic_clauses:
  - The Ghostfolio plugin must put the OpenFief transaction ID in each activity's comment.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 4f5acf388f1aeee5f9173cddc5071dea33373a9260cb82e96523bae954433f66
semantic_inventory:
  - claim_key: CLAIM-7F3B612B362716C7
    claim_text: The Ghostfolio plugin must put the OpenFief transaction ID in each activity's comment
    role: normative
    status: modeled
    span:
      start: 0
      end: 85
origin:
  kind: import
  ref: design-decisions:D-4#2
  recorded_at: '2026-10-27T20:12:55.615Z'
firmness: firm
required_evidence: asserted
id: REQ-AUTO-122F6AEE20C6FF6D
type: req
---
The Ghostfolio plugin must put the OpenFief transaction ID in each activity's comment.

## Source

> The Ghostfolio plugin runs a reconcile loop. It owns only the Ghostfolio accounts it created, puts OpenFief's transaction ID in each activity's comment, deletes stale activities and imports missing ones. Manual edits inside its accounts are reverted on the next sync.

Source: OpenFief author decisions (DESIGN.md section 10) - D-4#2

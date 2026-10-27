---
title: The Ghostfolio plugin must revert manual edits inside its accounts on the next sync.
status: open
text_ref: design-decisions:D-4#3
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-4-3
  - lane:strict
semantic_text: The Ghostfolio plugin must revert manual edits inside its accounts on the next sync.
logic_claims:
  - CLAIM-099AC20D86D56025
semantic_clauses:
  - The Ghostfolio plugin must revert manual edits inside its accounts on the next sync.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 494e2245e824123169dc616e4336a72039865510f2875092226518af572d5694
semantic_inventory:
  - claim_key: CLAIM-099AC20D86D56025
    claim_text: The Ghostfolio plugin must revert manual edits inside its accounts on the next sync
    role: normative
    status: modeled
    span:
      start: 0
      end: 83
origin:
  kind: import
  ref: design-decisions:D-4#3
  recorded_at: '2026-10-27T20:12:56.121Z'
firmness: firm
required_evidence: asserted
id: REQ-AUTO-4F8D1E2BADB7AED0
type: req
---
The Ghostfolio plugin must revert manual edits inside its accounts on the next sync.

## Source

> The Ghostfolio plugin runs a reconcile loop. It owns only the Ghostfolio accounts it created, puts OpenFief's transaction ID in each activity's comment, deletes stale activities and imports missing ones. Manual edits inside its accounts are reverted on the next sync.

Source: OpenFief author decisions (DESIGN.md section 10) - D-4#3

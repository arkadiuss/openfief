---
title: Broker position snapshots must be stored as balance assertions, and a mismatch must be flagged as an error.
status: open
text_ref: design-decisions:D-18#5
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-18-5
  - lane:strict
semantic_text: Broker position snapshots must be stored as balance assertions, and a mismatch must be flagged as an error.
logic_claims:
  - CLAIM-EDECF44B9E9AED2D
semantic_clauses:
  - Broker position snapshots must be stored as balance assertions, and a mismatch must be flagged as an error.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: d8419c52291f38887e647b37130a74b5dcf61d9e8beb97b10a52b082a52bc221
semantic_inventory:
  - claim_key: CLAIM-EDECF44B9E9AED2D
    claim_text: Broker position snapshots must be stored as balance assertions, and a mismatch must be flagged as an error
    role: normative
    status: modeled
    span:
      start: 0
      end: 106
origin:
  kind: import
  ref: design-decisions:D-18#5
  recorded_at: '2026-10-27T20:12:33.179Z'
firmness: firm
required_evidence: asserted
id: REQ-AUTO-D9FBE33256704D04
type: req
---
Broker position snapshots must be stored as balance assertions, and a mismatch must be flagged as an error.

## Source

> The source of truth is an append-only, versioned transaction log. Holdings and tax lots are derived from it. Re-imports are idempotent, keyed by source plus the institution's own ID; a changed row creates a new version and supersedes the old one, which is never deleted. Broker position snapshots are stored as balance assertions; a mismatch is a flagged error.

Source: OpenFief author decisions (DESIGN.md section 10) - D-18#5

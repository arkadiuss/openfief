---
title: The transaction log must be append-only and versioned.
status: open
text_ref: design-decisions:D-18#1
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-18-1
  - lane:strict
semantic_text: The transaction log must be append-only and versioned.
logic_claims:
  - CLAIM-524364D0F7D67F52
semantic_clauses:
  - The transaction log must be append-only and versioned.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 97cfa00190fdaa4735db34010a0e973788dfe91cc456ba278bfe8748fa8ca996
semantic_inventory:
  - claim_key: CLAIM-524364D0F7D67F52
    claim_text: The transaction log must be append-only and versioned
    role: normative
    status: modeled
    span:
      start: 0
      end: 53
origin:
  kind: import
  ref: design-decisions:D-18#1
  recorded_at: '2026-10-27T20:12:31.394Z'
firmness: firm
required_evidence: asserted
id: REQ-AUTO-CACBF41A8E14837D
type: req
---
The transaction log must be append-only and versioned.

## Source

> The source of truth is an append-only, versioned transaction log. Holdings and tax lots are derived from it. Re-imports are idempotent, keyed by source plus the institution's own ID; a changed row creates a new version and supersedes the old one, which is never deleted. Broker position snapshots are stored as balance assertions; a mismatch is a flagged error.

Source: OpenFief author decisions (DESIGN.md section 10) - D-18#1

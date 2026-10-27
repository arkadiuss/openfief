---
title: Re-imports must be idempotent, keyed by source plus the institution's own ID.
status: open
text_ref: design-decisions:D-18#3
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-18-3
  - lane:strict
semantic_text: Re-imports must be idempotent, keyed by source plus the institution's own ID.
logic_claims:
  - CLAIM-DBA5529AA7431C0E
semantic_clauses:
  - Re-imports must be idempotent, keyed by source plus the institution's own ID.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 467c763da0b09b49419cc795f3e21feddc69ba9716916bb3fd415a8511225566
semantic_inventory:
  - claim_key: CLAIM-DBA5529AA7431C0E
    claim_text: Re-imports must be idempotent, keyed by source plus the institution's own ID
    role: normative
    status: modeled
    span:
      start: 0
      end: 76
origin:
  kind: import
  ref: design-decisions:D-18#3
  recorded_at: '2026-10-27T20:12:32.242Z'
firmness: firm
required_evidence: asserted
id: REQ-AUTO-F19F621B67F14F4F
type: req
---
Re-imports must be idempotent, keyed by source plus the institution's own ID.

## Source

> The source of truth is an append-only, versioned transaction log. Holdings and tax lots are derived from it. Re-imports are idempotent, keyed by source plus the institution's own ID; a changed row creates a new version and supersedes the old one, which is never deleted. Broker position snapshots are stored as balance assertions; a mismatch is a flagged error.

Source: OpenFief author decisions (DESIGN.md section 10) - D-18#3

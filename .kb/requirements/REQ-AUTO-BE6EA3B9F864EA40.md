---
title: Holdings and tax lots must be derived from the transaction log.
status: open
text_ref: design-decisions:D-18#2
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-18-2
  - lane:strict
semantic_text: Holdings and tax lots must be derived from the transaction log.
logic_claims:
  - CLAIM-BD280B92907943D8
semantic_clauses:
  - Holdings and tax lots must be derived from the transaction log.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: e8260c2c516d0aa5e8a357f194fa152743fb4181ddc9045ebba2d737af4a148f
semantic_inventory:
  - claim_key: CLAIM-BD280B92907943D8
    claim_text: Holdings and tax lots must be derived from the transaction log
    role: normative
    status: modeled
    span:
      start: 0
      end: 62
origin:
  kind: import
  ref: design-decisions:D-18#2
  recorded_at: '2026-10-27T20:12:31.769Z'
firmness: firm
required_evidence: asserted
id: REQ-AUTO-BE6EA3B9F864EA40
type: req
---
Holdings and tax lots must be derived from the transaction log.

## Source

> The source of truth is an append-only, versioned transaction log. Holdings and tax lots are derived from it. Re-imports are idempotent, keyed by source plus the institution's own ID; a changed row creates a new version and supersedes the old one, which is never deleted. Broker position snapshots are stored as balance assertions; a mismatch is a flagged error.

Source: OpenFief author decisions (DESIGN.md section 10) - D-18#2

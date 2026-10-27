---
title: Scripts must authenticate with separate, scoped, expiring, read-only personal tokens.
status: open
text_ref: design-decisions:D-27#2
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-27-2
  - lane:strict
semantic_text: Scripts must authenticate with separate, scoped, expiring, read-only personal tokens.
logic_claims:
  - CLAIM-7283019D4F6C0AE0
semantic_clauses:
  - Scripts must authenticate with separate, scoped, expiring, read-only personal tokens.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 514f9483529ce89bb8664c6a6d8c2c8bc21f8f75bb6f0b1949b3d57223543c93
semantic_inventory:
  - claim_key: CLAIM-7283019D4F6C0AE0
    claim_text: Scripts must authenticate with separate, scoped, expiring, read-only personal tokens
    role: normative
    status: modeled
    span:
      start: 0
      end: 84
origin:
  kind: import
  ref: design-decisions:D-27#2
  recorded_at: '2026-10-27T20:12:45.078Z'
firmness: firm
required_evidence: asserted
id: REQ-AUTO-2D0C3CC0151C3C5C
type: req
---
Scripts must authenticate with separate, scoped, expiring, read-only personal tokens.

## Source

> Plugins authenticate to the core with projected ServiceAccount tokens (audience `openfief-core`). Scripts get separate, scoped, expiring, read-only personal tokens.

Source: OpenFief author decisions (DESIGN.md section 10) - D-27#2

---
title: Re-importing the same file must add nothing, and a changed row must create a new version while the old one remains, marked superseded.
status: open
text_ref: design-decisions:D-200#3
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-200-3
  - lane:strict
semantic_text: Re-importing the same file must add nothing, and a changed row must create a new version while the old one remains, marked superseded.
logic_claims:
  - CLAIM-16A8D2D3F13F3B32
semantic_clauses:
  - Re-importing the same file must add nothing, and a changed row must create a new version while the old one remains, marked superseded.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: c509771a8629c187c22e0ef5f895145ce723828c5b1c133e609a3de161234a58
semantic_inventory:
  - claim_key: CLAIM-16A8D2D3F13F3B32
    claim_text: Re-importing the same file must add nothing, and a changed row must create a new version while the old one remains, marked superseded
    role: normative
    status: modeled
    span:
      start: 0
      end: 133
firmness: firm
required_evidence: asserted
origin:
  kind: import
  ref: design-decisions:D-200#3
  recorded_at: '2026-11-07T03:31:36.127Z'
id: REQ-m0-exit-03-idempotent-versioned-imports
type: req
---
Re-importing the same file must add nothing, and a changed row must create a new version while the old one remains, marked superseded.

## Context

Author decision D-200: M0 is done when every one of its nineteen criteria holds. Each criterion is an end-to-end scenario on the firm requirement it checks, and the whole set passes on both kind and k3d (criterion 15); the scenarios that verify this criterion carry the tag acceptance:p10-3.

## Source

> **Idempotent, versioned imports.** Re-importing the same file adds nothing. A changed row creates a new version, and the old one remains, marked superseded.

design-decisions:D-200#3 (docs/DESIGN.md, D-200, criterion 3)

---
title: Every transaction's postings must sum to zero per currency, and a check over the whole log must pass.
status: open
text_ref: design-decisions:D-200#2
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-200-2
  - lane:strict
semantic_text: Every transaction's postings must sum to zero per currency, and a check over the whole log must pass.
logic_claims:
  - CLAIM-D9A7C6AD28106E36
semantic_clauses:
  - Every transaction's postings must sum to zero per currency, and a check over the whole log must pass.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 040f3e5be7f42fcb53501b1b20a68f677a01e95373c6ecca70700d4bdb982b50
semantic_inventory:
  - claim_key: CLAIM-D9A7C6AD28106E36
    claim_text: Every transaction's postings must sum to zero per currency, and a check over the whole log must pass
    role: normative
    status: modeled
    span:
      start: 0
      end: 100
firmness: firm
required_evidence: asserted
origin:
  kind: import
  ref: design-decisions:D-200#2
  recorded_at: '2026-11-07T03:31:31.427Z'
id: REQ-m0-exit-02-double-entry
type: req
---
Every transaction's postings must sum to zero per currency, and a check over the whole log must pass.

## Context

Author decision D-200: M0 is done when every one of its nineteen criteria holds. Each criterion is an end-to-end scenario on the firm requirement it checks, and the whole set passes on both kind and k3d (criterion 15); the scenarios that verify this criterion carry the tag acceptance:p10-2.

## Source

> **Double-entry.** Every transaction's postings sum to zero per currency, and a check over the whole log passes.

design-decisions:D-200#2 (docs/DESIGN.md, D-200, criterion 2)

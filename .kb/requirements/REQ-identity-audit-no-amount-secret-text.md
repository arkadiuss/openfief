---
title: An audit entry must never contain an amount, a secret or a question's text.
status: open
tags:
  - strict-modeling
  - lane:strict
  - provenance:design-decisions-d-211-4
text_ref: design-decisions:D-211#4
origin:
  kind: import
  ref: design-decisions:D-211#4
  recorded_at: '2026-11-12T19:18:17.908Z'
semantic_text: An audit entry must never contain an amount, a secret or a question's text.
logic_claims:
  - CLAIM-B38332D13D677E1D
semantic_clauses:
  - An audit entry must never contain an amount, a secret or a question's text.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
firmness: firm
required_evidence: asserted
semantic_source_hash: 39c118a32c41e7232a23f8070b847ccf1c5cdfa4dbaa78796663a1efa79c133e
semantic_inventory:
  - claim_key: CLAIM-B38332D13D677E1D
    claim_text: An audit entry must never contain an amount, a secret or a question's text
    role: normative
    status: modeled
    span:
      start: 0
      end: 74
id: REQ-identity-audit-no-amount-secret-text
type: req
---
An audit entry must never contain an amount, a secret or a question's text.

## Context

The author decided it with D-211 and GR-12, the privacy and security model. The built entry has no free-text field: every value is a member of a closed set, a time, or an ID of a fixed shape, and the table's checks repeat that. No end-to-end test proves it, so its evidence floor is asserted.

## Source

> Each entry names who, what, which records by ID, when and the outcome, never an amount, a secret or a question's text (GR-12).

OpenFief author decisions (docs/DESIGN.md), design-decisions:D-211#4
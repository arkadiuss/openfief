---
title: IBKR and XTB accounts must be seeded from current lot-level positions, each lot with its open date and cost, and its PLN cost at the NBP table A rate from the day before the open date.
status: open
text_ref: design-decisions:D-200#4
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-200-4
  - lane:strict
semantic_text: IBKR and XTB accounts must be seeded from current lot-level positions, each lot with its open date and cost, and its PLN cost at the NBP table A rate from the day before the open date.
logic_claims:
  - CLAIM-B8711C4562BA80F7
semantic_clauses:
  - IBKR and XTB accounts must be seeded from current lot-level positions, each lot with its open date and cost, and its PLN cost at the NBP table A rate from the day before the open date.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 3cdfc0e61179ff550d40439ac9b9a89db6da7b5239becfd4bbd8dd69b2c6b165
semantic_inventory:
  - claim_key: CLAIM-B8711C4562BA80F7
    claim_text: IBKR and XTB accounts must be seeded from current lot-level positions, each lot with its open date and cost, and its PLN cost at the NBP table A rate from the day before the open date
    role: normative
    status: modeled
    span:
      start: 0
      end: 183
firmness: firm
required_evidence: asserted
origin:
  kind: import
  ref: design-decisions:D-200#4
  recorded_at: '2026-11-07T03:31:40.772Z'
id: REQ-m0-exit-04-opening-lots
type: req
---
IBKR and XTB accounts must be seeded from current lot-level positions, each lot with its open date and cost, and its PLN cost at the NBP table A rate from the day before the open date.

## Context

Author decision D-200: M0 is done when every one of its nineteen criteria holds. Each criterion is an end-to-end scenario on the firm requirement it checks, and the whole set passes on both kind and k3d (criterion 15); the scenarios that verify this criterion carry the tag acceptance:p10-4.

## Source

> **Opening lots.** IBKR and XTB accounts are seeded from current lot-level positions, each lot with open date and cost, and PLN cost at the NBP table A rate from the day before the open date.

design-decisions:D-200#4 (docs/DESIGN.md, D-200, criterion 4)

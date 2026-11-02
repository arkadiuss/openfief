---
title: The account page must list the cash checkpoints the owner entered with their differences, offer the owner a form to enter one and a button to book a differing one's difference, and the attention item of a mismatched checkpoint must link to them.
status: open
text_ref: docs/DESIGN.md#D-158
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-158
  - lane:strict
semantic_text: The account page must list the cash checkpoints the owner entered with their differences, offer the owner a form to enter one and a button to book a differing one's difference, and the attention item of a mismatched checkpoint must link to them.
logic_claims:
  - CLAIM-3E0D01DDD5F2C1EE
semantic_clauses:
  - The account page must list the cash checkpoints the owner entered with their differences, offer the owner a form to enter one and a button to book a differing one's difference, and the attention item of a mismatched checkpoint must link to them.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: a50b6f948f92520104778a6e2884fa9f2ff3f507fc0cfc0d2df8e957756f7150
semantic_inventory:
  - claim_key: CLAIM-3E0D01DDD5F2C1EE
    claim_text: The account page must list the cash checkpoints the owner entered with their differences, offer the owner a form to enter one and a button to book a differing one's difference, and the attention item of a mismatched checkpoint must link to them
    role: normative
    status: modeled
    span:
      start: 0
      end: 244
origin:
  kind: agent
  recorded_at: '2026-11-02T03:50:26.454Z'
id: REQ-ui-cash-checkpoints
type: req
---
The account page must list the cash checkpoints the owner entered with their differences, offer the owner a form to enter one and a button to book a differing one's difference, and the attention item of a mismatched checkpoint must link to them.

## Context

D-158 says the owner enters the cash and books the difference from the attention item; docs/dev/design-system.md sets how the built-in UI shows it. Asked by the author through D-158; the page layout is this implementation's choice.

## Source

> The owner can enter the cash balance an institution shows for an account on a date, such as the bond registry's or a bank's. It is stored as a balance assertion (D-18), and a mismatch with the derived cash raises an attention item (D-130). From it the owner can book the difference as an ordinary manual transaction: a deposit or withdrawal by default, so an unexplained difference never shows as return, or cash income (D-152) when it is income or a fee. Nothing overwrites the derived balance.

docs/DESIGN.md#D-158

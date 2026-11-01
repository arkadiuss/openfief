---
title: The core must delete an account that holds no transactions or balance assertions, with its bindings, broker account mappings, tax wrapper, fee rule and stored valuations, for the owner only, and refuse one that holds any without changing anything.
status: open
text_ref: docs/DESIGN.md#D-120
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-120
  - lane:strict
semantic_text: The core must delete an account that holds no transactions or balance assertions, with its bindings, broker account mappings, tax wrapper, fee rule and stored valuations, for the owner only, and refuse one that holds any without changing anything.
logic_claims:
  - CLAIM-241F650A010D5FC0
semantic_clauses:
  - The core must delete an account that holds no transactions or balance assertions, with its bindings, broker account mappings, tax wrapper, fee rule and stored valuations, for the owner only, and refuse one that holds any without changing anything.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 4b768f2ff85e68ca8be7f6b36fab90153f990db0905f6fdf1cab33639f3c7517
semantic_inventory:
  - claim_key: CLAIM-241F650A010D5FC0
    claim_text: The core must delete an account that holds no transactions or balance assertions, with its bindings, broker account mappings, tax wrapper, fee rule and stored valuations, for the owner only, and refuse one that holds any without changing anything
    role: normative
    status: modeled
    span:
      start: 0
      end: 246
origin:
  kind: agent
  recorded_at: '2026-10-31T14:35:21.107Z'
id: REQ-ingest-delete-unused-account
type: req
---
The core must delete an account that holds no transactions or balance assertions, with its bindings, broker account mappings, tax wrapper, fee rule and stored valuations, for the owner only, and refuse one that holds any without changing anything.

## Context

D-120 lets the owner delete an account with no transactions; D-18 keeps everything the ledger holds. The ledger checks and deletes in one statement before the other contexts forget the account. Recorded by the implementation agent of impl/decisions-117; default firmness, not reviewed by the author yet.

## Source

> The owner can delete a sleeve with no accounts, a sleeve kind no sleeve uses and an account with no transactions, and remap a broker account whose OpenFief account has no transactions from it. Nothing the ledger holds is ever deleted (D-18).

docs/DESIGN.md#D-120

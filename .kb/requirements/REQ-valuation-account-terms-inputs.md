---
title: Recording an account's tax wrapper, one of the pack's, or its fee rule must revalue that account from its first transaction, the owner's marginal rate, above 0 and below 1, every account, and the owner's number format nothing, each refused without storing when invalid.
status: open
text_ref: docs/DESIGN.md#D-36
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-36
  - lane:strict
semantic_text: Recording an account's tax wrapper, one of the pack's, or its fee rule must revalue that account from its first transaction, the owner's marginal rate, above 0 and below 1, every account, and the owner's number format nothing, each refused without storing when invalid.
logic_claims:
  - CLAIM-71AD1EAC0D413F06
semantic_clauses:
  - Recording an account's tax wrapper, one of the pack's, or its fee rule must revalue that account from its first transaction, the owner's marginal rate, above 0 and below 1, every account, and the owner's number format nothing, each refused without storing when invalid.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 0a6724e4b5d19c70be49d378ea2f24e9bfbf0b32ce96449435c1a4444176604f
semantic_inventory:
  - claim_key: CLAIM-71AD1EAC0D413F06
    claim_text: Recording an account's tax wrapper, one of the pack's, or its fee rule must revalue that account from its first transaction, the owner's marginal rate, above 0 and below 1, every account, and the owner's number format nothing, each refused without storing when invalid
    role: normative
    status: modeled
    span:
      start: 0
      end: 268
origin:
  kind: agent
  recorded_at: '2026-10-29T21:49:10.007Z'
id: REQ-valuation-account-terms-inputs
type: req
---
Recording an account's tax wrapper, one of the pack's, or its fee rule must revalue that account from its first transaction, the owner's marginal rate, above 0 and below 1, every account, and the owner's number format nothing, each refused without storing when invalid.

## Context

D-36 revalues on any change to the inputs; the wrapper, the fee rule and the marginal rate are inputs of the values (D-76, D-78), the number format only of display (D-113). Modeled by Claude Code for the values work package; the author has not reviewed it yet.

## Source

> **[D-36]** Valuations are stored rows derived from the transaction log, prices, FX rates and index readings, and can be rebuilt from them at any time. [...] Configurable: worker concurrency and the end-of-day schedule. Each stored valuation holds all three values of D-77. What triggers a revaluation is not configurable.

docs/DESIGN.md#D-36

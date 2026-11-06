---
title: Each transfer suggested between the owner's own accounts and neither confirmed nor rejected must raise one transfer-suggested attention item in the decisions group, tied to both legs' accounts and listed only to a caller who may view both, which must clear once the owner confirms or rejects the pair.
status: open
text_ref: docs/DESIGN.md#D-130
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-130
  - lane:strict
semantic_text: Each transfer suggested between the owner's own accounts and neither confirmed nor rejected must raise one transfer-suggested attention item in the decisions group, tied to both legs' accounts and listed only to a caller who may view both, which must clear once the owner confirms or rejects the pair.
logic_claims:
  - CLAIM-1C1C871B5D74D859
semantic_clauses:
  - Each transfer suggested between the owner's own accounts and neither confirmed nor rejected must raise one transfer-suggested attention item in the decisions group, tied to both legs' accounts and listed only to a caller who may view both, which must clear once the owner confirms or rejects the pair.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 7568924437f1e84931cf734d759c97a9266c26da74ce132723555fcd02f2fca7
semantic_inventory:
  - claim_key: CLAIM-1C1C871B5D74D859
    claim_text: Each transfer suggested between the owner's own accounts and neither confirmed nor rejected must raise one transfer-suggested attention item in the decisions group, tied to both legs' accounts and listed only to a caller who may view both, which must clear once the owner confirms or rejects the pair
    role: normative
    status: modeled
    span:
      start: 0
      end: 300
origin:
  kind: agent
  recorded_at: '2026-11-06T06:15:07.244Z'
id: REQ-attention-transfer-suggested-item
type: req
---
Each transfer suggested between the owner's own accounts and neither confirmed nor rejected must raise one transfer-suggested attention item in the decisions group, tied to both legs' accounts and listed only to a caller who may view both, which must clear once the owner confirms or rejects the pair.

## Context

The design session answered that a suggested transfer is an attention item (D-130), so the Inbox's count and the overview's Needs you include the transfers waiting for the owner (D-105). Tying it to both accounts and hiding it unless the caller may view both follows transfer-broken (D-131, D-163). Modeled by Claude Code for the UI v2 overview slice; the author has not reviewed the wording.

## Source

> **[D-105]** Transfers between the owner's own accounts are identified by suggestion and confirmation. ... The owner confirms or rejects each suggestion in the UI, and a confirmed link is stored in the ledger, append-only.

docs/DESIGN.md#D-105

> **[D-130]** Every check that needs the owner raises an attention item in one append-only store ... and the UI lists them in one place.

docs/DESIGN.md#D-130

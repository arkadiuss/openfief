---
title: The owner must be able to reject a confirmed transfer whose legs a correction broke without the legs being checked.
status: open
text_ref: docs/DESIGN.md#D-131
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-131
  - lane:strict
semantic_text: The owner must be able to reject a confirmed transfer whose legs a correction broke without the legs being checked.
logic_claims:
  - CLAIM-74AFD3A6BE986BCA
semantic_clauses:
  - The owner must be able to reject a confirmed transfer whose legs a correction broke without the legs being checked.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 21450888329ae7e9a1a21070e2db0b141fda400b3ed2db3f22bd9f5828a7c4a2
semantic_inventory:
  - claim_key: CLAIM-74AFD3A6BE986BCA
    claim_text: The owner must be able to reject a confirmed transfer whose legs a correction broke without the legs being checked
    role: normative
    status: modeled
    span:
      start: 0
      end: 114
origin:
  kind: agent
  recorded_at: '2026-10-31T18:32:45.678Z'
id: REQ-ledger-transfer-broken-reject
type: req
---
The owner must be able to reject a confirmed transfer whose legs a correction broke without the legs being checked.

## Context

The attention items implementation (impl/attention): D-131 has the owner confirm the corrected pair or reject it, and rejecting must work although the legs no longer match. Default firmness.

## Source

> A confirmed transfer whose legs no longer match after a correction (D-18) stops counting and raises an attention item (D-130), so the owner confirms the corrected pair or rejects it; performance never changes silently.

docs/DESIGN.md#D-131

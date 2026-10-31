---
title: A confirmed transfer whose legs no longer match after a correction must raise an attention item, so the owner confirms the corrected pair or rejects it.
status: open
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-131-2
  - lane:strict
text_ref: design-decisions:D-131#2
origin:
  kind: import
  ref: design-decisions:D-131#2
  recorded_at: '2026-10-31T14:34:38.000Z'
semantic_text: A confirmed transfer whose legs no longer match after a correction must raise an attention item, so the owner confirms the corrected pair or rejects it.
logic_claims:
  - CLAIM-0C626D39ED5F47D1
semantic_clauses:
  - A confirmed transfer whose legs no longer match after a correction must raise an attention item, so the owner confirms the corrected pair or rejects it.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
firmness: firm
required_evidence: asserted
semantic_source_hash: cf0b985879c0aaace65e18d314a9f36784e0df784b5203b1ae24acd88b7b1af4
semantic_inventory:
  - claim_key: CLAIM-0C626D39ED5F47D1
    claim_text: A confirmed transfer whose legs no longer match after a correction must raise an attention item, so the owner confirms the corrected pair or rejects it
    role: normative
    status: modeled
    span:
      start: 0
      end: 151
id: REQ-ledger-transfer-broken-attention-item
type: req
---
A confirmed transfer whose legs no longer match after a correction must raise an attention item, so the owner confirms the corrected pair or rejects it.

## Context

A broker can correct either leg of a confirmed transfer (D-18), and the link then stops counting (REQ-ledger-transfer-broken-stops-counting). The implementation raised this as question #17 on docs/implementation-questions.md; today both legs simply become cash flows again and the pair is no longer listed as confirmed, so performance changes without the owner being told. Its options included flagging a broken link for the owner like a flagged balance assertion. On 2026-10-31 the author chose that, through D-130's attention items (REQ-core-attention-item-store), and recorded it as D-131 so that performance never changes silently.

## Source

> A confirmed transfer whose legs no longer match after a correction (D-18) stops counting and raises an attention item (D-130), so the owner confirms the corrected pair or rejects it; performance never changes silently.

OpenFief author decisions (docs/DESIGN.md, section 3.4), design-decisions:D-131#2

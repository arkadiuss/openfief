---
title: A confirmed transfer whose legs no longer match after a correction must stop counting.
status: open
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-131-1
  - lane:strict
text_ref: design-decisions:D-131#1
origin:
  kind: import
  ref: design-decisions:D-131#1
  recorded_at: '2026-10-31T14:34:38.000Z'
semantic_text: A confirmed transfer whose legs no longer match after a correction must stop counting.
logic_claims:
  - CLAIM-F7E5F590061F6A11
semantic_clauses:
  - A confirmed transfer whose legs no longer match after a correction must stop counting.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
firmness: firm
required_evidence: asserted
semantic_source_hash: afc6c3595383dd63a0ef1f1f300438536555b80e515a2af3cc434ec5738732c8
semantic_inventory:
  - claim_key: CLAIM-F7E5F590061F6A11
    claim_text: A confirmed transfer whose legs no longer match after a correction must stop counting
    role: normative
    status: modeled
    span:
      start: 0
      end: 85
id: REQ-ledger-transfer-broken-stops-counting
type: req
---
A confirmed transfer whose legs no longer match after a correction must stop counting.

## Context

D-105 stores the owner's confirmed transfer link append-only, keyed by the two transactions (REQ-ledger-transfer-link-append-only), and only confirmed transfers are left out of the time-weighted return's cash flows (REQ-portfolio-twr-confirmed-transfers-only). A broker can later correct either leg (D-18), say its amount or date, so that the two are no longer one possible transfer. The implementation raised this as question #17 on docs/implementation-questions.md: the link stays stored but counts only while the legs' current versions still match, otherwise both legs are cash flows again. On 2026-10-31 the author confirmed that the link stops counting, recorded as D-131, and added an attention item for the owner (REQ-ledger-transfer-broken-attention-item) so performance never changes silently.

## Source

> A confirmed transfer whose legs no longer match after a correction (D-18) stops counting and raises an attention item (D-130), so the owner confirms the corrected pair or rejects it; performance never changes silently.

OpenFief author decisions (docs/DESIGN.md, section 3.4), design-decisions:D-131#1

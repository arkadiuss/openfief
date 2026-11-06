---
title: 'The read API must list recent activity newest first, a page at a time: completed import batches and syncs, failed sync attempts, coupons booked and corrections, the last two only from accounts the caller may view, never an import''s or a sync''s amounts.'
status: open
text_ref: docs/DESIGN.md#D-193
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-193
  - lane:strict
semantic_text: 'The read API must list recent activity newest first, a page at a time: completed import batches and syncs, failed sync attempts, coupons booked and corrections, the last two only from accounts the caller may view, never an import''s or a sync''s amounts.'
logic_claims:
  - CLAIM-4B096141784A7DC5
semantic_clauses:
  - 'The read API must list recent activity newest first, a page at a time: completed import batches and syncs, failed sync attempts, coupons booked and corrections, the last two only from accounts the caller may view, never an import''s or a sync''s amounts.'
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 9d098b2947641a30672baa22da12c83b4ef7b656b8c1ae68c87344296aa12282
semantic_inventory:
  - claim_key: CLAIM-4B096141784A7DC5
    claim_text: 'The read API must list recent activity newest first, a page at a time: completed import batches and syncs, failed sync attempts, coupons booked and corrections, the last two only from accounts the caller may view, never an import''s or a sync''s amounts'
    role: normative
    status: modeled
    span:
      start: 0
      end: 251
origin:
  kind: agent
  recorded_at: '2026-11-06T01:58:38.389Z'
id: REQ-views-activity
type: req
---
The read API must list recent activity newest first, a page at a time: completed import batches and syncs, failed sync attempts, coupons booked and corrections, the last two only from accounts the caller may view, never an import's or a sync's amounts.

## Context

D-193 ends the overview with recent activity: imports, syncs, coupons and corrections. A failed sync attempt is listed as its own entry so the owner sees it there too. D-165 has imports and syncs name only their instance. Modeled by Claude Code for the UI v2 read API; the author has not reviewed the wording.

## Source

> **[D-193]** The overview opens with one sentence of status (the date everything is valued through, what needs the owner, any failing sync) and the Ask box, then ... the top of the Inbox with its actions, the accounts grouped by sleeve, and recent activity: imports, syncs, coupons and corrections.

docs/DESIGN.md#D-193

> **[D-163]** Authorization is per account and relationship-based. Every read path and owner service of the core asks one authorization interface whether the caller may act on each account ...

docs/DESIGN.md#D-163

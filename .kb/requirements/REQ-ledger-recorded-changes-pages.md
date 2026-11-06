---
title: The ledger store must list the coupons booked and the corrections recorded, deletions included, newest first by when each version was recorded and then by row ID, each with its postings, and continue after a given one.
status: open
text_ref: docs/DESIGN.md#D-193
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-193
  - lane:strict
semantic_text: The ledger store must list the coupons booked and the corrections recorded, deletions included, newest first by when each version was recorded and then by row ID, each with its postings, and continue after a given one.
logic_claims:
  - CLAIM-D1F7FA64EC406161
semantic_clauses:
  - The ledger store must list the coupons booked and the corrections recorded, deletions included, newest first by when each version was recorded and then by row ID, each with its postings, and continue after a given one.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 78b7a987239223b650fcda8abad0c6c245b197fbd9fd8a34aa73011a1d39a4ef
semantic_inventory:
  - claim_key: CLAIM-D1F7FA64EC406161
    claim_text: The ledger store must list the coupons booked and the corrections recorded, deletions included, newest first by when each version was recorded and then by row ID, each with its postings, and continue after a given one
    role: normative
    status: modeled
    span:
      start: 0
      end: 217
origin:
  kind: agent
  recorded_at: '2026-11-06T01:12:26.326Z'
id: REQ-ledger-recorded-changes-pages
type: req
---
The ledger store must list the coupons booked and the corrections recorded, deletions included, newest first by when each version was recorded and then by row ID, each with its postings, and continue after a given one.

## Context

D-193 ends the overview with recent activity: imports, syncs, coupons and corrections. The read API's activity reads the coupons and corrections from here; corrections are D-18's versions and D-100's deletions, coupons D-101's interest. Modeled by Claude Code for the UI v2 read API; the author has not reviewed the wording.

## Source

> **[D-193]** The overview opens with ... and recent activity: imports, syncs, coupons and corrections.

docs/DESIGN.md#D-193

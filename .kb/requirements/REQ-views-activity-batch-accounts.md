---
title: An import or sync in the recent activity must name the accounts it booked into that the caller may view, and must be left out when it booked only into accounts the caller may not view.
status: open
text_ref: docs/DESIGN.md#D-140
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-140
  - lane:strict
semantic_text: An import or sync in the recent activity must name the accounts it booked into that the caller may view, and must be left out when it booked only into accounts the caller may not view.
logic_claims:
  - CLAIM-55BD02C103F77DB7
semantic_clauses:
  - An import or sync in the recent activity must name the accounts it booked into that the caller may view, and must be left out when it booked only into accounts the caller may not view.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 3be7274280494b2189bcd716e7dc3f770a00b6ee456ab0fb6a317eb6fc829068
semantic_inventory:
  - claim_key: CLAIM-55BD02C103F77DB7
    claim_text: An import or sync in the recent activity must name the accounts it booked into that the caller may view, and must be left out when it booked only into accounts the caller may not view
    role: normative
    status: modeled
    span:
      start: 0
      end: 183
origin:
  kind: agent
  recorded_at: '2026-11-06T03:43:08.201Z'
id: REQ-views-activity-batch-accounts
type: req
---
An import or sync in the recent activity must name the accounts it booked into that the caller may view, and must be left out when it booked only into accounts the caller may not view.

## Context

A statement covering several accounts books every mapped account (D-109, D-140); the design session asked that the upload's entry in the recent activity (D-193) name them next to its instance, filtered by account read access (D-163). Modeled by Claude Code for the UI v2 Inbox; the author has not reviewed the wording.

## Source

> **[D-140]** An upload of a statement covering several accounts runs in the mode chosen for the account it was uploaded for, which the owner can override: an opening upload seeds only the accounts not seeded yet, and a transactions upload books every mapped account.

docs/DESIGN.md#D-140

> **[D-163]** Authorization is per account and relationship-based. Every read path and owner service of the core asks one authorization interface whether the caller may act on each account.

docs/DESIGN.md#D-163

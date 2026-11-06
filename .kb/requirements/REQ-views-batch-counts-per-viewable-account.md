---
title: An activity entry or a sync attempt that names several accounts must list only the accounts the caller may view, each with its own counts, and a caller who may view only some of them must see the sums of their counts, never totals across accounts it may not view, nor the rows tied to no account.
status: open
text_ref: docs/DESIGN.md#D-199
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-199
  - lane:strict
semantic_text: An activity entry or a sync attempt that names several accounts must list only the accounts the caller may view, each with its own counts, and a caller who may view only some of them must see the sums of their counts, never totals across accounts it may not view, nor the rows tied to no account.
logic_claims:
  - CLAIM-F8784D5E9806C93C
semantic_clauses:
  - An activity entry or a sync attempt that names several accounts must list only the accounts the caller may view, each with its own counts, and a caller who may view only some of them must see the sums of their counts, never totals across accounts it may not view, nor the rows tied to no account.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 6774ff9bb368b9ba982de66613452ac175f80f000686bffa409ca0632d659d85
semantic_inventory:
  - claim_key: CLAIM-F8784D5E9806C93C
    claim_text: An activity entry or a sync attempt that names several accounts must list only the accounts the caller may view, each with its own counts, and a caller who may view only some of them must see the sums of their counts, never totals across accounts it may not view, nor the rows tied to no account
    role: normative
    status: modeled
    span:
      start: 0
      end: 295
origin:
  kind: agent
  recorded_at: '2026-11-06T11:51:21.121Z'
id: REQ-views-batch-counts-per-viewable-account
type: req
---
An activity entry or a sync attempt that names several accounts must list only the accounts the caller may view, each with its own counts, and a caller who may view only some of them must see the sums of their counts, never totals across accounts it may not view, nor the rows tied to no account.

## Context

The design session refined D-199 for the UI v2 slice 3: counts per visible account, not totals across hidden ones. Modeled by Claude Code for the UI v2 pages slice; the author has not reviewed the wording.

## Source

> **[D-199]** Once viewers exist (D-65), a viewer sees the syncs, sync attempts and import activity of the accounts they may view; anything tied to a connector instance and no account stays the owner's, as D-165 has for attention items.

docs/DESIGN.md#D-199

Design session refinement (UI v2 slice 3 brief): an activity entry or sync attempt that names several accounts lists only the accounts the caller may view, with counts per visible account, not totals across hidden ones.

---
title: The attention sweep must raise one item of kind connector-sync-failing per connector instance whose sync keeps failing, with the instance as subject and when the failures began as detail, so it stays acknowledged while the same failures go on, resolves after the next success and opens again for a new run of failures.
status: open
text_ref: docs/DESIGN.md#D-196
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-196
  - lane:strict
semantic_text: The attention sweep must raise one item of kind connector-sync-failing per connector instance whose sync keeps failing, with the instance as subject and when the failures began as detail, so it stays acknowledged while the same failures go on, resolves after the next success and opens again for a new run of failures.
logic_claims:
  - CLAIM-ECCBD4961D87307A
semantic_clauses:
  - The attention sweep must raise one item of kind connector-sync-failing per connector instance whose sync keeps failing, with the instance as subject and when the failures began as detail, so it stays acknowledged while the same failures go on, resolves after the next success and opens again for a new run of failures.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 12abb31dc3873d102af540918002852086d5bab0eb91883d0baa77700f183345
semantic_inventory:
  - claim_key: CLAIM-ECCBD4961D87307A
    claim_text: The attention sweep must raise one item of kind connector-sync-failing per connector instance whose sync keeps failing, with the instance as subject and when the failures began as detail, so it stays acknowledged while the same failures go on, resolves after the next success and opens again for a new run of failures
    role: normative
    status: modeled
    span:
      start: 0
      end: 317
origin:
  kind: agent
  recorded_at: '2026-11-06T01:26:55.647Z'
id: REQ-attention-sync-failing-check
type: req
---
The attention sweep must raise one item of kind connector-sync-failing per connector instance whose sync keeps failing, with the instance as subject and when the failures began as detail, so it stays acknowledged while the same failures go on, resolves after the next success and opens again for a new run of failures.

## Context

Refines REQ-attention-sync-keeps-failing-item. A detail that changed on every attempt would reopen an acknowledged item each time; the start of the run doesn't change until a success ends it. Modeled by Claude Code for the UI v2 read API; the author has not reviewed the wording.

## Source

> The Inbox lists open attention items (D-130) grouped by what the owner does about them: data coming in (tickers to confirm, unmapped broker accounts, failing syncs, import problems), numbers to check (balance mismatches, price differences) and decisions (transfers to confirm, missing settings). An item that one field or button settles, such as confirming an ISIN, acknowledging a price check, confirming a transfer or retrying a sync, is settled in place; any other links to its account or instrument. Acknowledged and done items sit behind their own views, the count shows in the navigation and on the overview, and an account's page shows its own items at the top. A connector sync that keeps failing raises an item that shows what happened.

OpenFief author decisions (docs/DESIGN.md, section 2), design-decisions:D-196

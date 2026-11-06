---
title: The read API must serve each connector instance's sync status, including whether it keeps failing, and its sync attempts newest first a page at a time, each with its error class and never the error's text.
status: open
text_ref: docs/DESIGN.md#D-196
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-196
  - lane:strict
semantic_text: The read API must serve each connector instance's sync status, including whether it keeps failing, and its sync attempts newest first a page at a time, each with its error class and never the error's text.
logic_claims:
  - CLAIM-9627A3C181F70030
semantic_clauses:
  - The read API must serve each connector instance's sync status, including whether it keeps failing, and its sync attempts newest first a page at a time, each with its error class and never the error's text.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: c10eeefd2b84c79aab6900b578121f23bebf6c35094bce52716808ae9900b484
semantic_inventory:
  - claim_key: CLAIM-9627A3C181F70030
    claim_text: The read API must serve each connector instance's sync status, including whether it keeps failing, and its sync attempts newest first a page at a time, each with its error class and never the error's text
    role: normative
    status: modeled
    span:
      start: 0
      end: 204
origin:
  kind: agent
  recorded_at: '2026-11-06T01:58:43.430Z'
id: REQ-views-sync-history
type: req
---
The read API must serve each connector instance's sync status, including whether it keeps failing, and its sync attempts newest first a page at a time, each with its error class and never the error's text.

## Context

D-196 has a sync that keeps failing raise an item that shows what happened; this is what it shows. D-193 puts any failing sync in the overview's status sentence. GR-12 keeps error text out. Modeled by Claude Code for the UI v2 read API; the author has not reviewed the wording.

## Source

> **[D-196]** The Inbox lists open attention items (D-130) grouped by what the owner does about them ... any other links to its account or instrument. Acknowledged and done items sit behind their own views, the count shows in the navigation and on the overview, and an account's page shows its own items at the top. A connector sync that keeps failing raises an item that shows what happened.

docs/DESIGN.md#D-196

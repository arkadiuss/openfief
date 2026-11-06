---
title: The read API must list attention items by view (open, acknowledged or done) and by account, ISIN or connector instance, each with its group and scope, count the open, acknowledged and done items the filters select, and leave out an item tied to an account the caller may not view.
status: open
text_ref: docs/DESIGN.md#D-196
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-196
  - lane:strict
semantic_text: The read API must list attention items by view (open, acknowledged or done) and by account, ISIN or connector instance, each with its group and scope, count the open, acknowledged and done items the filters select, and leave out an item tied to an account the caller may not view.
logic_claims:
  - CLAIM-738581D03517F495
semantic_clauses:
  - The read API must list attention items by view (open, acknowledged or done) and by account, ISIN or connector instance, each with its group and scope, count the open, acknowledged and done items the filters select, and leave out an item tied to an account the caller may not view.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: e6c6ddfbddefab0df039d09524c35d5d5c6e7d67658da87821cc8498fa3fb9d5
semantic_inventory:
  - claim_key: CLAIM-738581D03517F495
    claim_text: The read API must list attention items by view (open, acknowledged or done) and by account, ISIN or connector instance, each with its group and scope, count the open, acknowledged and done items the filters select, and leave out an item tied to an account the caller may not view
    role: normative
    status: modeled
    span:
      start: 0
      end: 279
origin:
  kind: agent
  recorded_at: '2026-11-06T01:58:33.174Z'
id: REQ-views-attention-inbox
type: req
---
The read API must list attention items by view (open, acknowledged or done) and by account, ISIN or connector instance, each with its group and scope, count the open, acknowledged and done items the filters select, and leave out an item tied to an account the caller may not view.

## Context

Refines REQ-views-attention-items for D-196's Inbox: acknowledged and done items behind their own views, the open count in the navigation, an account's own items on its page, links to an item's account or instrument. D-163 limits each caller to the accounts it may view. Modeled by Claude Code for the UI v2 read API; the author has not reviewed the wording.

## Source

> **[D-196]** The Inbox lists open attention items (D-130) grouped by what the owner does about them ... any other links to its account or instrument. Acknowledged and done items sit behind their own views, the count shows in the navigation and on the overview, and an account's page shows its own items at the top. A connector sync that keeps failing raises an item that shows what happened.

docs/DESIGN.md#D-196

> **[D-163]** Authorization is per account and relationship-based. Every read path and owner service of the core asks one authorization interface whether the caller may act on each account ...

docs/DESIGN.md#D-163

---
title: A sync that keeps failing must show in the Inbox with its latest attempts, their error classes, the last good sync and the next try, and offer the owner Retry now.
status: open
text_ref: design-decisions:D-196
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-196
  - lane:strict
semantic_text: A sync that keeps failing must show in the Inbox with its latest attempts, their error classes, the last good sync and the next try, and offer the owner Retry now.
logic_claims:
  - CLAIM-0C57D3B992A0354F
semantic_clauses:
  - A sync that keeps failing must show in the Inbox with its latest attempts, their error classes, the last good sync and the next try, and offer the owner Retry now.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 02893995e3594ea4c3d39681fabd4cab31f39c938a2e5d7edaf79332ec678325
semantic_inventory:
  - claim_key: CLAIM-0C57D3B992A0354F
    claim_text: A sync that keeps failing must show in the Inbox with its latest attempts, their error classes, the last good sync and the next try, and offer the owner Retry now
    role: normative
    status: modeled
    span:
      start: 0
      end: 162
origin:
  kind: agent
  recorded_at: '2026-11-06T04:33:04.084Z'
id: REQ-ui-inbox-sync-timeline
type: req
---
A sync that keeps failing must show in the Inbox with its latest attempts, their error classes, the last good sync and the next try, and offer the owner Retry now.

## Context

D-196: the item shows what happened, and retrying a sync is settled in place. Modeled by Claude Code for the UI v2 Inbox; the author has not reviewed the wording.

## Source

> **[D-196]** The Inbox lists open attention items (D-130) grouped by what the owner does about them ... An item that one field or button settles, such as confirming an ISIN, acknowledging a price check, confirming a transfer or retrying a sync, is settled in place; any other links to its account or instrument. Acknowledged and done items sit behind their own views, the count shows in the navigation and on the overview, and an account's page shows its own items at the top. A connector sync that keeps failing raises an item that shows what happened.

OpenFief author decisions (docs/DESIGN.md), design-decisions:D-196

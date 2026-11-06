---
title: The owner must confirm a ticker's ISIN in the Inbox, with the proposal filled in or a field to type, and a trading currency choice when the ticker's exchange quotes in more than one, required when the proposal names none, and the Inbox must say how many waiting rows the confirmation booked.
status: open
text_ref: design-decisions:D-196
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-196
  - lane:strict
semantic_text: The owner must confirm a ticker's ISIN in the Inbox, with the proposal filled in or a field to type, and a trading currency choice when the ticker's exchange quotes in more than one, required when the proposal names none, and the Inbox must say how many waiting rows the confirmation booked.
logic_claims:
  - CLAIM-B1BF826322D9808F
semantic_clauses:
  - The owner must confirm a ticker's ISIN in the Inbox, with the proposal filled in or a field to type, and a trading currency choice when the ticker's exchange quotes in more than one, required when the proposal names none, and the Inbox must say how many waiting rows the confirmation booked.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 98754147bad93474333ee0b678f32aa105028735821ab2b095b6f066f8e2cd08
semantic_inventory:
  - claim_key: CLAIM-B1BF826322D9808F
    claim_text: The owner must confirm a ticker's ISIN in the Inbox, with the proposal filled in or a field to type, and a trading currency choice when the ticker's exchange quotes in more than one, required when the proposal names none, and the Inbox must say how many waiting rows the confirmation booked
    role: normative
    status: modeled
    span:
      start: 0
      end: 290
origin:
  kind: agent
  recorded_at: '2026-11-06T04:32:45.549Z'
id: REQ-ui-inbox-tickers
type: req
---
The owner must confirm a ticker's ISIN in the Inbox, with the proposal filled in or a field to type, and a trading currency choice when the ticker's exchange quotes in more than one, required when the proposal names none, and the Inbox must say how many waiting rows the confirmation booked.

## Context

D-116 and D-170 rules, moved from the tickers page to the Inbox (D-192, D-196). Modeled by Claude Code for the UI v2 Inbox; the author has not reviewed the wording.

## Source

> **[D-196]** The Inbox lists open attention items (D-130) grouped by what the owner does about them ... An item that one field or button settles, such as confirming an ISIN, acknowledging a price check, confirming a transfer or retrying a sync, is settled in place; any other links to its account or instrument. Acknowledged and done items sit behind their own views, the count shows in the navigation and on the overview, and an account's page shows its own items at the top. A connector sync that keeps failing raises an item that shows what happened.

OpenFief author decisions (docs/DESIGN.md), design-decisions:D-196

## Why this changed

D-192 removes the check pages (attention, tickers, balance assertions, price checks, transfers) and D-196 gathers what needs the owner in the Inbox, settled in place where one field or button does it; this requirement restates the old page's behavior for the Inbox.

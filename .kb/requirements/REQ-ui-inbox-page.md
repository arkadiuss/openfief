---
title: The built-in UI's Inbox must list attention items by view, open, acknowledged or done, each with its count, in cards for data coming in, numbers to check and decisions, each item with its title, detail, links to its accounts, status pill, history and the owner's actions, and must give a read session no forms.
status: open
text_ref: design-decisions:D-196
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-196
  - lane:strict
semantic_text: The built-in UI's Inbox must list attention items by view, open, acknowledged or done, each with its count, in cards for data coming in, numbers to check and decisions, each item with its title, detail, links to its accounts, status pill, history and the owner's actions, and must give a read session no forms.
logic_claims:
  - CLAIM-F00D972CC60FEB30
semantic_clauses:
  - The built-in UI's Inbox must list attention items by view, open, acknowledged or done, each with its count, in cards for data coming in, numbers to check and decisions, each item with its title, detail, links to its accounts, status pill, history and the owner's actions, and must give a read session no forms.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 81d14e9844a37abe39b235b4529441c22ca9c93efe29e53acc79a297f89f3400
semantic_inventory:
  - claim_key: CLAIM-F00D972CC60FEB30
    claim_text: The built-in UI's Inbox must list attention items by view, open, acknowledged or done, each with its count, in cards for data coming in, numbers to check and decisions, each item with its title, detail, links to its accounts, status pill, history and the owner's actions, and must give a read session no forms
    role: normative
    status: modeled
    span:
      start: 0
      end: 309
origin:
  kind: agent
  recorded_at: '2026-11-06T04:32:36.592Z'
id: REQ-ui-inbox-page
type: req
---
The built-in UI's Inbox must list attention items by view, open, acknowledged or done, each with its count, in cards for data coming in, numbers to check and decisions, each item with its title, detail, links to its accounts, status pill, history and the owner's actions, and must give a read session no forms.

## Context

The Inbox replaces the Attention page (D-192, D-196); the cards, pills and links follow the design system's M0 Inbox. Modeled by Claude Code for the UI v2 Inbox; the author has not reviewed the wording.

## Source

> **[D-196]** The Inbox lists open attention items (D-130) grouped by what the owner does about them ... An item that one field or button settles, such as confirming an ISIN, acknowledging a price check, confirming a transfer or retrying a sync, is settled in place; any other links to its account or instrument. Acknowledged and done items sit behind their own views, the count shows in the navigation and on the overview, and an account's page shows its own items at the top. A connector sync that keeps failing raises an item that shows what happened.

OpenFief author decisions (docs/DESIGN.md), design-decisions:D-196

## Why this changed

D-192 removes the check pages (attention, tickers, balance assertions, price checks, transfers) and D-196 gathers what needs the owner in the Inbox, settled in place where one field or button does it; this requirement restates the old page's behavior for the Inbox.

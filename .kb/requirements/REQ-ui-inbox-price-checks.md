---
title: A flagged price check's Inbox item must show both prices side by side with their sources, the difference, the threshold and the check's history, with the owner's acknowledgement and threshold form, and the done view must list every instrument's checks with their history, for any session.
status: open
text_ref: design-decisions:D-196
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-196
  - lane:strict
semantic_text: A flagged price check's Inbox item must show both prices side by side with their sources, the difference, the threshold and the check's history, with the owner's acknowledgement and threshold form, and the done view must list every instrument's checks with their history, for any session.
logic_claims:
  - CLAIM-89875598F8D18125
semantic_clauses:
  - A flagged price check's Inbox item must show both prices side by side with their sources, the difference, the threshold and the check's history, with the owner's acknowledgement and threshold form, and the done view must list every instrument's checks with their history, for any session.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: c61f96dd65368b04d5c72935a476df35df55a92312e51285258ce45cd4b06bbc
semantic_inventory:
  - claim_key: CLAIM-89875598F8D18125
    claim_text: A flagged price check's Inbox item must show both prices side by side with their sources, the difference, the threshold and the check's history, with the owner's acknowledgement and threshold form, and the done view must list every instrument's checks with their history, for any session
    role: normative
    status: modeled
    span:
      start: 0
      end: 287
origin:
  kind: agent
  recorded_at: '2026-11-06T04:32:56.855Z'
id: REQ-ui-inbox-price-checks
type: req
---
A flagged price check's Inbox item must show both prices side by side with their sources, the difference, the threshold and the check's history, with the owner's acknowledgement and threshold form, and the done view must list every instrument's checks with their history, for any session.

## Context

D-110 and D-112, moved from the price checks page to the Inbox (D-192, D-196). Modeled by Claude Code for the UI v2 Inbox; the author has not reviewed the wording.

## Source

> **[D-196]** The Inbox lists open attention items (D-130) grouped by what the owner does about them ... An item that one field or button settles, such as confirming an ISIN, acknowledging a price check, confirming a transfer or retrying a sync, is settled in place; any other links to its account or instrument. Acknowledged and done items sit behind their own views, the count shows in the navigation and on the overview, and an account's page shows its own items at the top. A connector sync that keeps failing raises an item that shows what happened.

OpenFief author decisions (docs/DESIGN.md), design-decisions:D-196

## Why this changed

D-192 removes the check pages (attention, tickers, balance assertions, price checks, transfers) and D-196 gathers what needs the owner in the Inbox, settled in place where one field or button does it; this requirement restates the old page's behavior for the Inbox.

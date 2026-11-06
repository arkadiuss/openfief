---
title: The Inbox must list the suggested transfers among the decisions for the owner to confirm or mark as not a transfer, and its done view the confirmed and rejected pairs with the decision that corrects each.
status: open
text_ref: design-decisions:D-196
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-196
  - lane:strict
semantic_text: The Inbox must list the suggested transfers among the decisions for the owner to confirm or mark as not a transfer, and its done view the confirmed and rejected pairs with the decision that corrects each.
logic_claims:
  - CLAIM-95D5E937C9FBA9F7
semantic_clauses:
  - The Inbox must list the suggested transfers among the decisions for the owner to confirm or mark as not a transfer, and its done view the confirmed and rejected pairs with the decision that corrects each.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: bfa1fae8eff94015e0f7f635d69f11b699dbfecd2de41ca48e301fc9b44c5b01
semantic_inventory:
  - claim_key: CLAIM-95D5E937C9FBA9F7
    claim_text: The Inbox must list the suggested transfers among the decisions for the owner to confirm or mark as not a transfer, and its done view the confirmed and rejected pairs with the decision that corrects each
    role: normative
    status: modeled
    span:
      start: 0
      end: 203
origin:
  kind: agent
  recorded_at: '2026-11-06T04:32:51.629Z'
id: REQ-ui-inbox-transfers
type: req
---
The Inbox must list the suggested transfers among the decisions for the owner to confirm or mark as not a transfer, and its done view the confirmed and rejected pairs with the decision that corrects each.

## Context

D-105's decisions, moved from the transfers page to the Inbox (D-192, D-196). Suggested transfers are not attention items, so they are listed beside the items and not counted. Modeled by Claude Code for the UI v2 Inbox; the author has not reviewed the wording.

## Source

> **[D-196]** The Inbox lists open attention items (D-130) grouped by what the owner does about them ... An item that one field or button settles, such as confirming an ISIN, acknowledging a price check, confirming a transfer or retrying a sync, is settled in place; any other links to its account or instrument. Acknowledged and done items sit behind their own views, the count shows in the navigation and on the overview, and an account's page shows its own items at the top. A connector sync that keeps failing raises an item that shows what happened.

OpenFief author decisions (docs/DESIGN.md), design-decisions:D-196

## Why this changed

D-192 removes the check pages (attention, tickers, balance assertions, price checks, transfers) and D-196 gathers what needs the owner in the Inbox, settled in place where one field or button does it; this requirement restates the old page's behavior for the Inbox.

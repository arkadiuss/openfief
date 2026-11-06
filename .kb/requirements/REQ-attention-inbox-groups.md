---
title: 'Open attention items must be grouped by what the owner does about them: data coming in (tickers to confirm, unmapped broker accounts, failing syncs, import problems), numbers to check (balance mismatches, price differences) and decisions (transfers to confirm, missing settings).'
status: open
text_ref: design-decisions:D-196#1
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-196-1
  - lane:strict
semantic_text: 'Open attention items must be grouped by what the owner does about them: data coming in (tickers to confirm, unmapped broker accounts, failing syncs, import problems), numbers to check (balance mismatches, price differences) and decisions (transfers to confirm, missing settings).'
logic_claims:
  - CLAIM-F883E74817667C31
semantic_clauses:
  - 'Open attention items must be grouped by what the owner does about them: data coming in (tickers to confirm, unmapped broker accounts, failing syncs, import problems), numbers to check (balance mismatches, price differences) and decisions (transfers to confirm, missing settings).'
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 19e3c8e65df50d8b87c39a07cc58a5f41e2efe477879731fc8feaea85668d99d
semantic_inventory:
  - claim_key: CLAIM-F883E74817667C31
    claim_text: 'Open attention items must be grouped by what the owner does about them: data coming in (tickers to confirm, unmapped broker accounts, failing syncs, import problems), numbers to check (balance mismatches, price differences) and decisions (transfers to confirm, missing settings)'
    role: normative
    status: modeled
    span:
      start: 0
      end: 278
firmness: firm
required_evidence: asserted
origin:
  kind: import
  ref: design-decisions:D-196#1
  recorded_at: '2026-11-06T01:26:50.424Z'
id: REQ-attention-inbox-groups
type: req
---
Open attention items must be grouped by what the owner does about them: data coming in (tickers to confirm, unmapped broker accounts, failing syncs, import problems), numbers to check (balance mismatches, price differences) and decisions (transfers to confirm, missing settings).

## Context

The author decided D-196 with the UI around an Inbox (D-192 to D-196). The read API gives each item its group so every client groups alike; REQ-attention-group-of-kind places the kinds the decision doesn't name.

## Source

> The Inbox lists open attention items (D-130) grouped by what the owner does about them: data coming in (tickers to confirm, unmapped broker accounts, failing syncs, import problems), numbers to check (balance mismatches, price differences) and decisions (transfers to confirm, missing settings). An item that one field or button settles, such as confirming an ISIN, acknowledging a price check, confirming a transfer or retrying a sync, is settled in place; any other links to its account or instrument. Acknowledged and done items sit behind their own views, the count shows in the navigation and on the overview, and an account's page shows its own items at the top. A connector sync that keeps failing raises an item that shows what happened.

OpenFief author decisions (docs/DESIGN.md, section 2), design-decisions:D-196

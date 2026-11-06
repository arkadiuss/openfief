---
title: Acknowledged and done attention items must sit behind their own views, the open count must show in the navigation and on the overview, and an account's page must show its own items at the top.
status: open
text_ref: design-decisions:D-196#3
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-196-3
  - lane:strict
semantic_text: Acknowledged and done attention items must sit behind their own views, the open count must show in the navigation and on the overview, and an account's page must show its own items at the top.
logic_claims:
  - CLAIM-1712807651932E9C
semantic_clauses:
  - Acknowledged and done attention items must sit behind their own views, the open count must show in the navigation and on the overview, and an account's page must show its own items at the top.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 5e3dbc172c78cf776134e128a5ff60e3a7f87cfe32bae384b403a74356dfc6e6
semantic_inventory:
  - claim_key: CLAIM-1712807651932E9C
    claim_text: Acknowledged and done attention items must sit behind their own views, the open count must show in the navigation and on the overview, and an account's page must show its own items at the top
    role: normative
    status: modeled
    span:
      start: 0
      end: 191
firmness: firm
required_evidence: asserted
origin:
  kind: import
  ref: design-decisions:D-196#3
  recorded_at: '2026-11-06T04:32:29.424Z'
id: REQ-ui-inbox-views-and-counts
type: req
---
Acknowledged and done attention items must sit behind their own views, the open count must show in the navigation and on the overview, and an account's page must show its own items at the top.

## Context

Stated by the author in D-196.

## Source

> **[D-196]** The Inbox lists open attention items (D-130) grouped by what the owner does about them ... An item that one field or button settles, such as confirming an ISIN, acknowledging a price check, confirming a transfer or retrying a sync, is settled in place; any other links to its account or instrument. Acknowledged and done items sit behind their own views, the count shows in the navigation and on the overview, and an account's page shows its own items at the top. A connector sync that keeps failing raises an item that shows what happened.

OpenFief author decisions (docs/DESIGN.md), design-decisions:D-196

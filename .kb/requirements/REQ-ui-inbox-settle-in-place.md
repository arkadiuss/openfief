---
title: An attention item that one field or button settles, such as confirming an ISIN, acknowledging a price check, confirming a transfer or retrying a sync, must be settled in place in the Inbox; any other must link to its account or instrument.
status: open
text_ref: design-decisions:D-196#2
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-196-2
  - lane:strict
semantic_text: An attention item that one field or button settles, such as confirming an ISIN, acknowledging a price check, confirming a transfer or retrying a sync, must be settled in place in the Inbox; any other must link to its account or instrument.
logic_claims:
  - CLAIM-5EE63BECC3AB750F
semantic_clauses:
  - An attention item that one field or button settles, such as confirming an ISIN, acknowledging a price check, confirming a transfer or retrying a sync, must be settled in place in the Inbox; any other must link to its account or instrument.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: bf52da5cb3287fccc22634bfb55a67077dbfc47b6d8a8673393a96a21da95b21
semantic_inventory:
  - claim_key: CLAIM-5EE63BECC3AB750F
    claim_text: An attention item that one field or button settles, such as confirming an ISIN, acknowledging a price check, confirming a transfer or retrying a sync, must be settled in place in the Inbox; any other must link to its account or instrument
    role: normative
    status: modeled
    span:
      start: 0
      end: 238
firmness: firm
required_evidence: asserted
origin:
  kind: import
  ref: design-decisions:D-196#2
  recorded_at: '2026-11-06T04:32:25.836Z'
id: REQ-ui-inbox-settle-in-place
type: req
---
An attention item that one field or button settles, such as confirming an ISIN, acknowledging a price check, confirming a transfer or retrying a sync, must be settled in place in the Inbox; any other must link to its account or instrument.

## Context

Stated by the author in D-196.

## Source

> **[D-196]** The Inbox lists open attention items (D-130) grouped by what the owner does about them ... An item that one field or button settles, such as confirming an ISIN, acknowledging a price check, confirming a transfer or retrying a sync, is settled in place; any other links to its account or instrument. Acknowledged and done items sit behind their own views, the count shows in the navigation and on the overview, and an account's page shows its own items at the top. A connector sync that keeps failing raises an item that shows what happened.

OpenFief author decisions (docs/DESIGN.md), design-decisions:D-196

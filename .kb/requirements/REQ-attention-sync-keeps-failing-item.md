---
title: A connector sync that keeps failing must raise an attention item that shows what happened.
status: open
text_ref: design-decisions:D-196#4
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-196-4
  - lane:strict
semantic_text: A connector sync that keeps failing must raise an attention item that shows what happened.
logic_claims:
  - CLAIM-909B3D3B41E774A3
semantic_clauses:
  - A connector sync that keeps failing must raise an attention item that shows what happened.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 391dfa20cc68524214428dde65c1e4d24e0d230f672f53af770cd7e07af63dee
semantic_inventory:
  - claim_key: CLAIM-909B3D3B41E774A3
    claim_text: A connector sync that keeps failing must raise an attention item that shows what happened
    role: normative
    status: modeled
    span:
      start: 0
      end: 89
firmness: firm
required_evidence: asserted
origin:
  kind: import
  ref: design-decisions:D-196#4
  recorded_at: '2026-11-06T01:26:44.522Z'
id: REQ-attention-sync-keeps-failing-item
type: req
---
A connector sync that keeps failing must raise an attention item that shows what happened.

## Context

The author decided D-196 with the UI around an Inbox (D-192 to D-196). The item shows its timeline through the read API's sync history; when a sync keeps failing is REQ-ingest-sync-keeps-failing.

## Source

> The Inbox lists open attention items (D-130) grouped by what the owner does about them: data coming in (tickers to confirm, unmapped broker accounts, failing syncs, import problems), numbers to check (balance mismatches, price differences) and decisions (transfers to confirm, missing settings). An item that one field or button settles, such as confirming an ISIN, acknowledging a price check, confirming a transfer or retrying a sync, is settled in place; any other links to its account or instrument. Acknowledged and done items sit behind their own views, the count shows in the navigation and on the overview, and an account's page shows its own items at the top. A connector sync that keeps failing raises an item that shows what happened.

OpenFief author decisions (docs/DESIGN.md, section 2), design-decisions:D-196

---
title: Every attention item must carry the account, the instrument and the connector instance its kind and subject name, each empty when none, and an item tied to no account must name only its connector instance.
status: open
text_ref: docs/DESIGN.md#D-165
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-165
  - lane:strict
semantic_text: Every attention item must carry the account, the instrument and the connector instance its kind and subject name, each empty when none, and an item tied to no account must name only its connector instance.
logic_claims:
  - CLAIM-E9D12E275F5CE83D
semantic_clauses:
  - Every attention item must carry the account, the instrument and the connector instance its kind and subject name, each empty when none, and an item tied to no account must name only its connector instance.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 8100d39edd38a023ee677e396527d0d444501d0f975aab27d7b3346a7d6f6669
semantic_inventory:
  - claim_key: CLAIM-E9D12E275F5CE83D
    claim_text: Every attention item must carry the account, the instrument and the connector instance its kind and subject name, each empty when none, and an item tied to no account must name only its connector instance
    role: normative
    status: modeled
    span:
      start: 0
      end: 204
origin:
  kind: agent
  recorded_at: '2026-11-06T01:27:06.623Z'
id: REQ-attention-item-scope
type: req
---
Every attention item must carry the account, the instrument and the connector instance its kind and subject name, each empty when none, and an item tied to no account must name only its connector instance.

## Context

D-196 links an item that isn't settled in place to its account or instrument and shows an account's own items on its page; D-165 has an item tied to no account name only its instance; D-163 hides an item tied to an account the caller may not view. Modeled by Claude Code for the UI v2 read API; the author has not reviewed the wording.

## Source

> The Inbox lists open attention items (D-130) grouped by what the owner does about them: data coming in (tickers to confirm, unmapped broker accounts, failing syncs, import problems), numbers to check (balance mismatches, price differences) and decisions (transfers to confirm, missing settings). An item that one field or button settles, such as confirming an ISIN, acknowledging a price check, confirming a transfer or retrying a sync, is settled in place; any other links to its account or instrument. Acknowledged and done items sit behind their own views, the count shows in the navigation and on the overview, and an account's page shows its own items at the top. A connector sync that keeps failing raises an item that shows what happened.

OpenFief author decisions (docs/DESIGN.md, section 2), design-decisions:D-196

> **[D-165]** An attention item about something tied to no account, such as an unmapped broker account, names only its connector instance on the read API. Once viewers exist (D-65), only the owner sees such items.

docs/DESIGN.md#D-165

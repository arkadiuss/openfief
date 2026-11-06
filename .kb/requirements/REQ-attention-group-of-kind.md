---
title: 'Attention items of kinds D-196 doesn''t name must be grouped too: a market-data gap and dropped undeclared deliveries as data coming in, an instrument to classify and any kind the core doesn''t know as decisions.'
status: open
text_ref: docs/DESIGN.md#D-196
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-196
  - lane:strict
semantic_text: 'Attention items of kinds D-196 doesn''t name must be grouped too: a market-data gap and dropped undeclared deliveries as data coming in, an instrument to classify and any kind the core doesn''t know as decisions.'
logic_claims:
  - CLAIM-0D9A4528FB0649A2
semantic_clauses:
  - 'Attention items of kinds D-196 doesn''t name must be grouped too: a market-data gap and dropped undeclared deliveries as data coming in, an instrument to classify and any kind the core doesn''t know as decisions.'
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 9467316cc5a782b3cd230405ac64e109e2e15f327bac62d04785b8856ebefa83
semantic_inventory:
  - claim_key: CLAIM-0D9A4528FB0649A2
    claim_text: 'Attention items of kinds D-196 doesn''t name must be grouped too: a market-data gap and dropped undeclared deliveries as data coming in, an instrument to classify and any kind the core doesn''t know as decisions'
    role: normative
    status: modeled
    span:
      start: 0
      end: 209
origin:
  kind: agent
  recorded_at: '2026-11-06T01:27:00.962Z'
id: REQ-attention-group-of-kind
type: req
---
Attention items of kinds D-196 doesn't name must be grouped too: a market-data gap and dropped undeclared deliveries as data coming in, an instrument to classify and any kind the core doesn't know as decisions.

## Context

Refines REQ-attention-inbox-groups for the kinds other contexts raise (D-94, D-132, D-179, D-186). An unknown kind is a decision because only the owner can say what it means. Modeled by Claude Code for the UI v2 read API; the author has not reviewed the wording.

## Source

> The Inbox lists open attention items (D-130) grouped by what the owner does about them: data coming in (tickers to confirm, unmapped broker accounts, failing syncs, import problems), numbers to check (balance mismatches, price differences) and decisions (transfers to confirm, missing settings). An item that one field or button settles, such as confirming an ISIN, acknowledging a price check, confirming a transfer or retrying a sync, is settled in place; any other links to its account or instrument. Acknowledged and done items sit behind their own views, the count shows in the navigation and on the overview, and an account's page shows its own items at the top. A connector sync that keeps failing raises an item that shows what happened.

OpenFief author decisions (docs/DESIGN.md, section 2), design-decisions:D-196

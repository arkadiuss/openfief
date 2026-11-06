---
title: 'A transfer-broken attention item must be tied to both legs'' accounts: the read API must name them, list the item under either account''s filter, and list it only to a caller who may view each of them.'
status: open
text_ref: docs/DESIGN.md#D-131
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-131
  - lane:strict
semantic_text: 'A transfer-broken attention item must be tied to both legs'' accounts: the read API must name them, list the item under either account''s filter, and list it only to a caller who may view each of them.'
logic_claims:
  - CLAIM-7289816A6BAD8EFB
semantic_clauses:
  - 'A transfer-broken attention item must be tied to both legs'' accounts: the read API must name them, list the item under either account''s filter, and list it only to a caller who may view each of them.'
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 717e365865956559708a213d5e437fc8ca46a403f7f9577b51e256a74f2dff60
semantic_inventory:
  - claim_key: CLAIM-7289816A6BAD8EFB
    claim_text: 'A transfer-broken attention item must be tied to both legs'' accounts: the read API must name them, list the item under either account''s filter, and list it only to a caller who may view each of them'
    role: normative
    status: modeled
    span:
      start: 0
      end: 198
origin:
  kind: agent
  recorded_at: '2026-11-06T03:43:03.750Z'
id: REQ-views-transfer-broken-accounts
type: req
---
A transfer-broken attention item must be tied to both legs' accounts: the read API must name them, list the item under either account's filter, and list it only to a caller who may view each of them.

## Context

The Inbox (D-196) shows an account's own items in its banner. A broken transfer (D-131) is about two accounts, so the design session asked that both accounts' banners show it, filtered by read access to the accounts (D-163); listing it only when the caller may view both matches how the transfers list hides a pair with a hidden leg. Modeled by Claude Code for the UI v2 Inbox; the author has not reviewed the wording.

## Source

> **[D-131]** ... a broken transfer link ... raises an attention item.

docs/DESIGN.md#D-131

> **[D-163]** Authorization is per account and relationship-based. Every read path and owner service of the core asks one authorization interface whether the caller may act on each account.

docs/DESIGN.md#D-163

---
title: A read API query about one account the caller may not view must fail exactly as one about an account that is not stored.
status: open
text_ref: docs/DESIGN.md#D-163
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-163
  - lane:strict
semantic_text: A read API query about one account the caller may not view must fail exactly as one about an account that is not stored.
logic_claims:
  - CLAIM-1E973B1F1CB11FE6
semantic_clauses:
  - A read API query about one account the caller may not view must fail exactly as one about an account that is not stored.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: b39edd065b1a1f6747333cac562afad6e0b8b364484c14b7aba86929c31c8320
semantic_inventory:
  - claim_key: CLAIM-1E973B1F1CB11FE6
    claim_text: A read API query about one account the caller may not view must fail exactly as one about an account that is not stored
    role: normative
    status: modeled
    span:
      start: 0
      end: 119
origin:
  kind: agent
  recorded_at: '2026-11-01T22:46:15.826Z'
id: REQ-views-authz-hidden-account-not-found
type: req
---
A read API query about one account the caller may not view must fail exactly as one about an account that is not stored.

## Context

Refines REQ-identity-authz-per-account. D-163 doesn't say how a refused account is answered; answering it as not found, with the same fixed text, keeps a caller from learning which accounts exist. Modeled by Claude Code for the authorization slice; the author has not reviewed it yet.

## Source

> Authorization is per account and relationship-based. Every read path and owner service of the core asks one authorization interface whether the caller may act on each account, and totals, the allocation and performance are computed over the accounts the caller may view. Until SpiceDB answers it (D-164), the interface's only implementation lets the owner act on every account and gives every other holder of read access (D-137) a view of every account.

docs/DESIGN.md#D-163

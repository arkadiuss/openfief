---
title: An owner service must refuse a change of an account the authorizer denies the owner, before it reads or changes anything about that account.
status: open
text_ref: docs/DESIGN.md#D-163
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-163
  - lane:strict
semantic_text: An owner service must refuse a change of an account the authorizer denies the owner, before it reads or changes anything about that account.
logic_claims:
  - CLAIM-30CF6FDFD2A2A3D8
semantic_clauses:
  - An owner service must refuse a change of an account the authorizer denies the owner, before it reads or changes anything about that account.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 5cd5d1d92d7779ad22dc6962e9a97f61a5773797903a7534562709963c568079
semantic_inventory:
  - claim_key: CLAIM-30CF6FDFD2A2A3D8
    claim_text: An owner service must refuse a change of an account the authorizer denies the owner, before it reads or changes anything about that account
    role: normative
    status: modeled
    span:
      start: 0
      end: 139
origin:
  kind: agent
  recorded_at: '2026-11-01T22:46:26.043Z'
id: REQ-core-authz-change-refused
type: req
---
An owner service must refuse a change of an account the authorizer denies the owner, before it reads or changes anything about that account.

## Context

Refines REQ-identity-authz-one-interface for the owner services (D-79). Asking first means a refused account looks the same whether it is stored or not; the refusal is the context's ErrNotOwner, permission denied over Connect. Modeled by Claude Code for the authorization slice; the author has not reviewed it yet.

## Source

> Authorization is per account and relationship-based. Every read path and owner service of the core asks one authorization interface whether the caller may act on each account, and totals, the allocation and performance are computed over the accounts the caller may view. Until SpiceDB answers it (D-164), the interface's only implementation lets the owner act on every account and gives every other holder of read access (D-137) a view of every account.

docs/DESIGN.md#D-163

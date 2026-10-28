---
title: Adding, subtracting or comparing money in different currencies must fail with a currency mismatch error.
status: open
text_ref: docs/DESIGN.md#D-20
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-20
  - lane:strict
semantic_text: Adding, subtracting or comparing money in different currencies must fail with a currency mismatch error.
logic_claims:
  - CLAIM-9DF2BBDA9FB14F3E
semantic_clauses:
  - Adding, subtracting or comparing money in different currencies must fail with a currency mismatch error.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 8d554dc3c2ae7ebd0789cb7984346cb4c5f979b1afe163a6bc62a6b4fcf31bbc
semantic_inventory:
  - claim_key: CLAIM-9DF2BBDA9FB14F3E
    claim_text: Adding, subtracting or comparing money in different currencies must fail with a currency mismatch error
    role: normative
    status: modeled
    span:
      start: 0
      end: 103
origin:
  kind: agent
  recorded_at: '2026-10-28T04:48:38.347Z'
id: REQ-kernel-money-currency-mismatch
type: req
---
Adding, subtracting or comparing money in different currencies must fail with a currency mismatch error.

## Context

Refines the firm requirement REQ-AUTO-5CB834BAA25E2C4C (amounts always keep their original currency, D-20) for the shared kernel's Money type. Kernel arithmetic never converts, so a sum or comparison across currencies is a data or programming error that callers must see as a typed CurrencyMismatchError, not as a silently wrong number. Conversion to PLN happens only in valuation, with an explicit FX rate. Modeled by Claude Code for the M0 foundation task (shared kernel); the author has not reviewed it yet.

## Source

> Every amount keeps its original currency.

docs/DESIGN.md D-20

> Every amount carries its currency.

docs/dev/code-organization.md#rules (Money, P)

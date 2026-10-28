---
title: An account must be stored only if it is valid and its sleeve is stored, and transactions and balance assertions must be imported only into stored accounts.
status: open
text_ref: docs/DESIGN.md#D-18
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-18
  - lane:strict
semantic_text: An account must be stored only if it is valid and its sleeve is stored, and transactions and balance assertions must be imported only into stored accounts.
logic_claims:
  - CLAIM-373C61F59E8FFE72
semantic_clauses:
  - An account must be stored only if it is valid and its sleeve is stored, and transactions and balance assertions must be imported only into stored accounts.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: abced7404c276d9542013badfdec99fce6ec3ed466ce7cb0d52d82104ac0f56f
semantic_inventory:
  - claim_key: CLAIM-373C61F59E8FFE72
    claim_text: An account must be stored only if it is valid and its sleeve is stored, and transactions and balance assertions must be imported only into stored accounts
    role: normative
    status: modeled
    span:
      start: 0
      end: 154
origin:
  kind: agent
  recorded_at: '2026-10-28T06:06:00.503Z'
id: REQ-ledger-app-registry
type: req
---
An account must be stored only if it is valid and its sleeve is stored, and transactions and balance assertions must be imported only into stored accounts.

## Context

The glossary says an account belongs to one sleeve, and allocation by sleeve (D-30) needs every account's sleeve to exist. Refusing rows for unknown accounts keeps the log from holding transactions that no view can place. Sleeves are portfolio's concept; until the portfolio context has its own store, the ledger store keeps them next to the accounts that refer to them. Modeled by Claude Code for the M0 ledger-store task; the author has not reviewed it yet.

## Source

> Account: One account at an institution, such as one IBKR linked account. Belongs to one sleeve.

docs/glossary.md (Account)

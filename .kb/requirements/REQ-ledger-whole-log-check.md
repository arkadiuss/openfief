---
title: 'The owner must be able to check the whole transaction log: every stored version of every transaction, current or superseded, must be summed per currency, each version that does not sum to zero named, and the transactions, versions and superseded versions counted per account, with only counts logged.'
status: open
text_ref: docs/DESIGN.md#D-200
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-200
  - lane:strict
semantic_text: 'The owner must be able to check the whole transaction log: every stored version of every transaction, current or superseded, must be summed per currency, each version that does not sum to zero named, and the transactions, versions and superseded versions counted per account, with only counts logged.'
logic_claims:
  - CLAIM-A997EF84463D42AA
semantic_clauses:
  - 'The owner must be able to check the whole transaction log: every stored version of every transaction, current or superseded, must be summed per currency, each version that does not sum to zero named, and the transactions, versions and superseded versions counted per account, with only counts logged.'
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: eb05021466a8a6bc2a4c996085a0fb14e9faa4d9b15ffb62a90776b47a3c42ff
semantic_inventory:
  - claim_key: CLAIM-A997EF84463D42AA
    claim_text: 'The owner must be able to check the whole transaction log: every stored version of every transaction, current or superseded, must be summed per currency, each version that does not sum to zero named, and the transactions, versions and superseded versions counted per account, with only counts logged'
    role: normative
    status: modeled
    span:
      start: 0
      end: 299
origin:
  kind: agent
  recorded_at: '2026-11-07T04:21:52.154Z'
id: REQ-ledger-whole-log-check
type: req
---
The owner must be able to check the whole transaction log: every stored version of every transaction, current or superseded, must be summed per currency, each version that does not sum to zero named, and the transactions, versions and superseded versions counted per account, with only counts logged.

## Context

D-200 criterion 2 asks that every transaction postings sum to zero per currency and that a check over the whole log passes (D-19). The owner runs it as core check ledger in the core pod, as core check balances runs; the end-to-end data path test runs it after its imports.

## Source

> Double-entry. Every transaction postings sum to zero per currency, and a check over the whole log passes.

docs/DESIGN.md#D-200

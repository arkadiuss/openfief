---
title: A transaction must have at least two postings that sum to zero in every currency, and an unbalanced transaction must be rejected with the currency and amount it is off by.
status: open
text_ref: docs/DESIGN.md#D-19
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-19
  - lane:strict
semantic_text: A transaction must have at least two postings that sum to zero in every currency, and an unbalanced transaction must be rejected with the currency and amount it is off by.
logic_claims:
  - CLAIM-91ED7AB82979DBB9
semantic_clauses:
  - A transaction must have at least two postings that sum to zero in every currency, and an unbalanced transaction must be rejected with the currency and amount it is off by.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: e3545cc313676156d25adf9b5557dec44a0dabc78f01096836557ad22b9bc1a0
semantic_inventory:
  - claim_key: CLAIM-91ED7AB82979DBB9
    claim_text: A transaction must have at least two postings that sum to zero in every currency, and an unbalanced transaction must be rejected with the currency and amount it is off by
    role: normative
    status: modeled
    span:
      start: 0
      end: 170
origin:
  kind: agent
  recorded_at: '2026-10-28T05:24:45.791Z'
id: REQ-ledger-postings-zero-sum
type: req
---
A transaction must have at least two postings that sum to zero in every currency, and an unbalanced transaction must be rejected with the currency and amount it is off by.

## Context

Refines the firm requirement REQ-AUTO-8C3982CF80A9A4DF (D-19). The glossary defines a posting as one leg of a transaction whose postings sum to zero per currency. Sums are per currency because amounts keep their original currency (D-20) and are never converted inside the ledger; an FX conversion balances through one FX ledger account per currency. Naming the currency and difference lets a connector author find a mapping bug. Modeled by Claude Code for the M0 ledger task; the author has not reviewed it yet.

## Source

> Double-entry postings across cash, position, fee, withholding-tax and FX accounts.

docs/DESIGN.md D-19

> One leg of a transaction. A transaction's postings sum to zero per currency.

docs/glossary.md (Posting)

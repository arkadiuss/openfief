---
title: The built-in UI must show each account's holdings with their stored values, open lots, cash, balance assertion checks and transactions page by page, and list every flagged balance assertion check.
status: open
text_ref: docs/DESIGN.md#D-18
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-18
  - lane:strict
semantic_text: The built-in UI must show each account's holdings with their stored values, open lots, cash, balance assertion checks and transactions page by page, and list every flagged balance assertion check.
logic_claims:
  - CLAIM-42486D559F71AB49
semantic_clauses:
  - The built-in UI must show each account's holdings with their stored values, open lots, cash, balance assertion checks and transactions page by page, and list every flagged balance assertion check.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: fd7a01ae425a032dbb553ccac6cff7023bfeb45939664d2649cf4f890b50ada6
semantic_inventory:
  - claim_key: CLAIM-42486D559F71AB49
    claim_text: The built-in UI must show each account's holdings with their stored values, open lots, cash, balance assertion checks and transactions page by page, and list every flagged balance assertion check
    role: normative
    status: modeled
    span:
      start: 0
      end: 195
origin:
  kind: agent
  recorded_at: '2026-10-28T19:50:02.142Z'
id: REQ-ui-account-pages
type: req
---
The built-in UI must show each account's holdings with their stored values, open lots, cash, balance assertion checks and transactions page by page, and list every flagged balance assertion check.

## Context

D-18 (a balance assertion mismatch is a flagged error) and D-6 (the read API covers accounts, transactions, tax lots, holdings and valuations); the orchestrator's task lists the account pages.

## Source

> Broker position snapshots are stored as balance assertions; a mismatch is a flagged error.

docs/DESIGN.md#D-18

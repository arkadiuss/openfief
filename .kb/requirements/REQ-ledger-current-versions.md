---
title: Holdings, lots and balances must be derived only from the current version of each transaction, and superseded versions must never count toward them.
status: open
text_ref: docs/DESIGN.md#D-18
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-18
  - lane:strict
semantic_text: Holdings, lots and balances must be derived only from the current version of each transaction, and superseded versions must never count toward them.
logic_claims:
  - CLAIM-47DA94D67C39B413
semantic_clauses:
  - Holdings, lots and balances must be derived only from the current version of each transaction, and superseded versions must never count toward them.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 2611b141e1354f01c2a939d4cd344cf45cea44c1dc63b657e275fcb45fb7f355
semantic_inventory:
  - claim_key: CLAIM-47DA94D67C39B413
    claim_text: Holdings, lots and balances must be derived only from the current version of each transaction, and superseded versions must never count toward them
    role: normative
    status: modeled
    span:
      start: 0
      end: 147
origin:
  kind: agent
  recorded_at: '2026-10-28T05:32:34.421Z'
id: REQ-ledger-current-versions
type: req
---
Holdings, lots and balances must be derived only from the current version of each transaction, and superseded versions must never count toward them.

## Context

Refines the firm requirement REQ-AUTO-BE6EA3B9F864EA40 (D-18: holdings and tax lots are derived from the transaction log). Because superseded versions stay in the log, every derivation must first reduce the log to the current version per key, or a corrected row would be counted twice. Modeled by Claude Code for the M0 ledger task; the author has not reviewed it yet.

## Source

> Holdings and tax lots are derived from it.

docs/DESIGN.md D-18

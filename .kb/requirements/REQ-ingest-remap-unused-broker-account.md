---
title: The core must remap a broker account to another account or to none, for the owner only, while the account it is mapped to holds no transaction from its connector instance, and refuse it otherwise.
status: open
text_ref: docs/DESIGN.md#D-120
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-120
  - lane:strict
semantic_text: The core must remap a broker account to another account or to none, for the owner only, while the account it is mapped to holds no transaction from its connector instance, and refuse it otherwise.
logic_claims:
  - CLAIM-A9C4D60E1D91016C
semantic_clauses:
  - The core must remap a broker account to another account or to none, for the owner only, while the account it is mapped to holds no transaction from its connector instance, and refuse it otherwise.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 75c45a433375a53073656cca3add85cc09665abe149a83cfa867433246d05a57
semantic_inventory:
  - claim_key: CLAIM-A9C4D60E1D91016C
    claim_text: The core must remap a broker account to another account or to none, for the owner only, while the account it is mapped to holds no transaction from its connector instance, and refuse it otherwise
    role: normative
    status: modeled
    span:
      start: 0
      end: 195
origin:
  kind: agent
  recorded_at: '2026-10-31T14:35:25.599Z'
id: REQ-ingest-remap-unused-broker-account
type: req
---
The core must remap a broker account to another account or to none, for the owner only, while the account it is mapped to holds no transaction from its connector instance, and refuse it otherwise.

## Context

D-120. Transactions carry no broker account, so 'from it' is read as from its connector instance, whose key sources every such transaction has (implementation questions). Recorded by the implementation agent of impl/decisions-117; default firmness, not reviewed by the author yet.

## Source

> The owner can delete a sleeve with no accounts, a sleeve kind no sleeve uses and an account with no transactions, and remap a broker account whose OpenFief account has no transactions from it. Nothing the ledger holds is ever deleted (D-18).

docs/DESIGN.md#D-120

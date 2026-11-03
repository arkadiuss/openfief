---
title: Stores Only Valid Accounts = true
status: active
text_ref: docs/DESIGN.md#D-18
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-18
  - lane:strict
  - fact:property_value
fact_kind: property_value
subject_key: ledger.account_registry
property_key: stores_only_valid_accounts
canonical_key: mcp-kibi-model-requirement:ledger.account_registry:stores_only_valid_accounts:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-98ABCCC6310C9B66
claim_text: An account must be stored only if it is valid, and transactions and balance assertions must be imported only into stored accounts.
origin:
  kind: agent
  recorded_at: '2026-11-03T11:50:03.682Z'
id: FACT-PROP-A1DBF270A9183DA6
type: fact
---

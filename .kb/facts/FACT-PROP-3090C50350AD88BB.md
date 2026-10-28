---
title: Opening Cash Against Equity = true
status: active
text_ref: docs/DESIGN.md#D-33
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-33
  - lane:strict
  - fact:property_value
fact_kind: property_value
subject_key: ledger.accounts
property_key: opening_cash_against_equity
canonical_key: mcp-kibi-model-requirement:ledger.accounts:opening_cash_against_equity:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-00CCC2DC91FA8392
claim_text: Opening cash must be booked as a deposit, or a withdrawal for a negative balance, against the equity ledger account of its currency, keyed with the opening cash prefix so that the ledger can tell it from other deposits.
origin:
  kind: agent
  recorded_at: '2026-10-28T08:39:04.352Z'
id: FACT-PROP-3090C50350AD88BB
type: fact
---

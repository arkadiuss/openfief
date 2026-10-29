---
title: Excludes Transfers Between Own Accounts = true
status: active
text_ref: docs/DESIGN.md#D-61
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-61
  - lane:strict
  - fact:property_value
fact_kind: property_value
subject_key: views.performance
property_key: excludes_transfers_between_own_accounts
canonical_key: mcp-kibi-model-requirement:views.performance:excludes_transfers_between_own_accounts:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-21A98D574F5D9E1B
claim_text: A withdrawal from one account and a deposit into another account of exactly the same amount and currency, dated at most five days apart in either order, must be matched as a transfer that is not an external cash flow, each withdrawal taking the nearest unmatched deposit, and opening cash must never be a leg.
origin:
  kind: agent
  recorded_at: '2026-10-28T20:55:04.245Z'
id: FACT-PROP-ADF09B88F059E14E
type: fact
---

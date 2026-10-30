---
title: Names Broker Accounts = true
status: active
text_ref: docs/DESIGN.md#D-72
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-72
  - lane:strict
  - fact:property_value
fact_kind: property_value
subject_key: connector.contract
property_key: names_broker_accounts
canonical_key: mcp-kibi-model-requirement:connector.contract:names_broker_accounts:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-9251EB4C9F549979
claim_text: The connector contract must let a transaction and a balance assertion name a broker account instead of an OpenFief account, never both, and let a delivery list every broker account it saw, as additions only, and the core must refuse a delivery naming both or a malformed broker account.
origin:
  kind: agent
  recorded_at: '2026-10-29T11:37:05.304Z'
id: FACT-PROP-17646B667B6EDC0E
type: fact
---

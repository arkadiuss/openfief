---
title: Refuses Unbound Accounts = true
status: active
text_ref: docs/DESIGN.md#D-48
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-48
  - lane:strict
  - fact:property_value
fact_kind: property_value
subject_key: ingest.connector_binding
property_key: refuses_unbound_accounts
canonical_key: mcp-kibi-model-requirement:ingest.connector_binding:refuses_unbound_accounts:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-11CFCF371D6CF259
claim_text: A connector instance's delivery must be refused whole, before anything is imported, when any of its transactions or balance assertions is in an account the instance has no connector binding to, and a connector binding must only be added for a stored account and a valid instance name.
origin:
  kind: agent
  recorded_at: '2026-10-28T16:50:21.950Z'
id: FACT-PROP-FB2EB63BD50A2AE3
type: fact
---

---
title: Owner Account Entries = true
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
subject_key: ingest.account_service
property_key: owner_account_entries
canonical_key: mcp-kibi-model-requirement:ingest.account_service:owner_account_entries:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-005ADE0FE57798F1
claim_text: 'The core must serve the owner''s account entries as one service to view plugins acting for an owner session: creating an account with an ID made from its name, its institution, sleeve, the connector instance that feeds it and the broker account it is, binding the instance and mapping the broker account, changing an account''s name or sleeve, and binding an instance to an account, refusing a taken ID, an unknown sleeve or account, a malformed instance and a broker account the instance never reported or that is mapped to another account, before anything is stored.'
origin:
  kind: agent
  recorded_at: '2026-10-29T12:46:01.718Z'
id: FACT-PROP-7758BCB5DC0609CF
type: fact
---

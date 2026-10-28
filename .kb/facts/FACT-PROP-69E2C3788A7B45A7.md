---
title: Logs Request Ids Only = true
status: active
text_ref: docs/dev/code-organization.md#secrets-and-data
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-dev-code-organization-md-secrets-and-data
  - lane:strict
  - fact:property_value
fact_kind: property_value
subject_key: views.read_api
property_key: logs_request_ids_only
canonical_key: mcp-kibi-model-requirement:views.read_api:logs_request_ids_only:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-7739A99185AEA8C0
claim_text: The read API must log requests only by an internal request ID, method, status and failing step, never a path, amount, holding, account identifier, token or error text, and must set no-store, nosniff, frame-denying, no-referrer and restrictive content-security headers on every response.
origin:
  kind: agent
  recorded_at: '2026-10-28T08:44:24.761Z'
id: FACT-PROP-69E2C3788A7B45A7
type: fact
---

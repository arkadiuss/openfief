---
title: Owner Only = true
status: active
text_ref: docs/DESIGN.md#D-200
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-200
  - lane:strict
  - fact:property_value
fact_kind: property_value
subject_key: views.balance_check
property_key: owner_only
canonical_key: mcp-kibi-model-requirement:views.balance_check:owner_only:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-91551CAD1022A290
claim_text: 'The balance report must be served to the owner alone: the views must refuse it to every other caller, such as a script token, a view plugin or a demo session, no read API method may serve it, and the CLI must reach it by running the core''s check command in the core''s pod through Kubernetes exec with the owner''s kubeconfig, over the accounts the authorizer lets the owner view.'
origin:
  kind: agent
  recorded_at: '2026-11-06T23:37:21.988Z'
id: FACT-PROP-0C0FA0B1BD429EC4
type: fact
---

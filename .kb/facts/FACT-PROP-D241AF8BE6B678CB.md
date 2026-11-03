---
title: Requested Secret Is Instance Double Hyphen Role = true
status: active
text_ref: docs/DESIGN.md#D-89
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-89
  - lane:strict
  - fact:property_value
fact_kind: property_value
subject_key: plugins.secret_roles
property_key: requested_secret_is_instance_double_hyphen_role
canonical_key: mcp-kibi-model-requirement:plugins.secret_roles:requested_secret_is_instance_double_hyphen_role:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-C97E4F8383CA0D82
claim_text: A plugin instance must request, and its workload must mount, the Secret <instance>--<role> for each Secret role its manifest names, and a role must be a DNS label without two hyphens in a row.
origin:
  kind: agent
  recorded_at: '2026-11-03T18:24:24.955Z'
id: FACT-PROP-D241AF8BE6B678CB
type: fact
---

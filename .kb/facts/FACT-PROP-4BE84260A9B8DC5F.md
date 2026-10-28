---
title: Is Request Grant Intersection = true
status: active
text_ref: docs/DESIGN.md#D-16
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-16
  - lane:strict
  - fact:property_value
fact_kind: property_value
subject_key: plugins.effective_permissions
property_key: is_request_grant_intersection
canonical_key: mcp-kibi-model-requirement:plugins.effective_permissions:is_request_grant_intersection:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-C1AF41DBEE088861
claim_text: A plugin's effective permissions must be exactly the capabilities, egress hosts, Secrets and data tiers that both its Plugin requests and its Grant approves, and a plugin with no approved capability must get none.
origin:
  kind: agent
  recorded_at: '2026-10-28T05:28:18.450Z'
id: FACT-PROP-4BE84260A9B8DC5F
type: fact
---

---
title: Instance Secret Name From Role = true
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
property_key: instance_secret_name_from_role
canonical_key: mcp-kibi-model-requirement:plugins.secret_roles:instance_secret_name_from_role:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-CB4583049B0C1862
claim_text: A plugin instance must request, and its workload must mount, the Secret <instance>-<role> for each Secret role its manifest names, and a role must be a DNS label.
origin:
  kind: agent
  recorded_at: '2026-10-29T15:10:28.954Z'
id: FACT-PROP-559EE0ED4036581A
type: fact
---

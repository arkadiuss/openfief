---
title: Mounted Secret Name = <instance>-<role>
status: active
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-89-2
  - lane:strict
  - fact:property_value
text_ref: design-decisions:D-89#2
fact_kind: property_value
subject_key: operator.plugin_objects
property_key: mounted_secret_name
operator: eq
value_type: string
value_string: <instance>-<role>
canonical_key: docs-design-md:operator.plugin_objects:mounted_secret_name:eq:instance_-_role
claim_key: CLAIM-46766ECC6EC662A2
claim_text: For each Secret role in a plugin's manifest, the operator must mount the Secret named <instance>-<role>.
origin:
  kind: agent
  recorded_at: '2026-10-29T08:32:48.819Z'
id: FACT-PROP-9F1E4CCDC47D6D40
type: fact
---

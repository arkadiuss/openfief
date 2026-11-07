---
title: Uninstall Removes Instance Keeps Data = true
status: active
text_ref: design-decisions:D-200#18
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-200-18
  - lane:strict
  - fact:property_value
fact_kind: property_value
subject_key: m0.exit
property_key: uninstall_removes_instance_keeps_data
canonical_key: mcp-kibi-model-requirement:m0.exit:uninstall_removes_instance_keeps_data:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-E3DC34EF944717D0
claim_text: Uninstalling a connector instance must remove its pod, grant, secrets and allow-list, while its imported transactions remain, marked as coming from a removed instance.
origin:
  kind: agent
  recorded_at: '2026-11-07T03:32:48.296Z'
id: FACT-PROP-F973A06B113EA359
type: fact
---

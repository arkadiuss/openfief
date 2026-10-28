---
title: Removes Runtime And Keeps Data = true
status: active
text_ref: design-decisions:D-58#1
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-58-1
  - lane:strict
  - fact:property_value
fact_kind: property_value
subject_key: plugins.uninstall
property_key: removes_runtime_and_keeps_data
canonical_key: docs-design-md:plugins.uninstall:removes_runtime_and_keeps_data:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-2DFAAB34A91C1117
claim_text: Uninstalling a plugin instance must delete its workload, grant, secrets and egress allow-list, and must keep the data it brought in, marked as coming from a removed instance.
origin:
  kind: agent
  recorded_at: '2026-10-28T06:09:34.129Z'
id: FACT-PROP-AF7A7A36B38036B2
type: fact
---

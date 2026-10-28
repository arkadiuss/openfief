---
title: Refuses All Without Allow List = true
status: active
text_ref: design-decisions:D-63#1
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-63-1
  - lane:strict
  - fact:property_value
fact_kind: property_value
subject_key: plugins.egress_proxy
property_key: refuses_all_without_allow_list
canonical_key: docs-design-md:plugins.egress_proxy:refuses_all_without_allow_list:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-84EE04E0241DDA78
claim_text: The egress proxy must never run without an allow-list and must refuse every destination for a plugin that has none.
origin:
  kind: agent
  recorded_at: '2026-10-28T06:18:44.317Z'
id: FACT-PROP-8937EBA1D6F57526
type: fact
---

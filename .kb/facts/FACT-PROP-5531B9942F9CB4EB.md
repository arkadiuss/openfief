---
title: Matches Exact Host And Port = true
status: active
text_ref: docs/DESIGN.md#D-32
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-32
  - lane:strict
  - fact:property_value
fact_kind: property_value
subject_key: plugins.egress_allow_list
property_key: matches_exact_host_and_port
canonical_key: mcp-kibi-model-requirement:plugins.egress_allow_list:matches_exact_host_and_port:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-CF486E213DCBB024
claim_text: The egress proxy must allow a destination only when its host and port exactly match an entry in the plugin's egress allow-list, ignoring only ASCII case and one trailing dot.
origin:
  kind: agent
  recorded_at: '2026-10-28T05:28:46.594Z'
id: FACT-PROP-5531B9942F9CB4EB
type: fact
---

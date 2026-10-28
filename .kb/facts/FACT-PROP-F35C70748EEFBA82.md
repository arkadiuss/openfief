---
title: Swapped Atomically = true
status: active
text_ref: docs/DESIGN.md#D-63
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-63
  - lane:strict
  - fact:property_value
fact_kind: property_value
subject_key: plugins.egress_allow_list
property_key: swapped_atomically
canonical_key: mcp-kibi-model-requirement:plugins.egress_allow_list:swapped_atomically:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-33F1F60CA967B7EA
claim_text: A changed egress allow-list must replace the previous one atomically, so each CONNECT is decided against one complete list and the next CONNECT after a change sees the new list.
origin:
  kind: agent
  recorded_at: '2026-10-28T08:47:13.529Z'
id: FACT-PROP-F35C70748EEFBA82
type: fact
---

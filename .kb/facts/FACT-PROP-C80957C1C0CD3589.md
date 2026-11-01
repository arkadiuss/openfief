---
title: Records Refusal Event = true
status: active
text_ref: docs/DESIGN.md#D-56
tags:
  - strict-modeling
  - confidence:0.85
  - confidence-band:medium
  - provenance:docs-design-md-d-56
  - lane:strict
  - fact:property_value
fact_kind: property_value
subject_key: operator.image_signature
property_key: records_refusal_event
canonical_key: mcp-kibi-model-requirement:operator.image_signature:records_refusal_event:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-B31ADC2A8D96AE6D
claim_text: When a reconcile newly refuses a Plugin image signature, the operator must record one Warning event on the Plugin naming the refusal reason and message, and must not record it again while the refusal is unchanged.
origin:
  kind: agent
  recorded_at: '2026-11-01T13:51:12.240Z'
id: FACT-PROP-C80957C1C0CD3589
type: fact
---

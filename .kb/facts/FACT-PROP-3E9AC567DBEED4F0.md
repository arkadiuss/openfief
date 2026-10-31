---
title: Serves Uploads Only = true
status: active
text_ref: docs/DESIGN.md#D-9
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-9
  - lane:strict
  - fact:property_value
fact_kind: property_value
subject_key: plugins.xtb
property_key: serves_uploads_only
canonical_key: mcp-kibi-model-requirement:plugins.xtb:serves_uploads_only:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-5357CCCD71EB167B
claim_text: The xtb plugin must serve ImportFile through the SDK with the XTB mapper, deliver nothing on Sync, and ask in its manifest for statement files only, with no egress and no Secret.
origin:
  kind: agent
  recorded_at: '2026-10-31T12:56:10.312Z'
id: FACT-PROP-3E9AC567DBEED4F0
type: fact
---

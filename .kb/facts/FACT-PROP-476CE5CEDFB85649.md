---
title: Holds No Request Contents = true
status: active
text_ref: docs/DESIGN.md#D-200
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-200
  - lane:strict
  - fact:property_value
fact_kind: property_value
subject_key: e2e.egress_log
property_key: holds_no_request_contents
canonical_key: mcp-kibi-model-requirement:e2e.egress_log:holds_no_request_contents:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-66A05045510D7407
claim_text: On both kind and k3d, the egress proxy's log must show each plugin request with its host, its decision and the plugin instance, and nothing of its path, query, headers, body or token.
origin:
  kind: agent
  recorded_at: '2026-11-07T03:31:11.910Z'
id: FACT-PROP-476CE5CEDFB85649
type: fact
---

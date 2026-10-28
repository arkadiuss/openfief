---
title: Has Strict Configmap Format = true
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
property_key: has_strict_configmap_format
canonical_key: mcp-kibi-model-requirement:plugins.egress_allow_list:has_strict_configmap_format:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-139EE916C1FD4853
claim_text: The operator must hand egress allow-lists to the egress proxy as ConfigMap openfief-egress-allowlist with one key per plugin whose value lists canonical host:port lines, and the proxy must refuse data that is not in exactly that form.
origin:
  kind: agent
  recorded_at: '2026-10-28T06:05:54.028Z'
id: FACT-PROP-B217939AE1EB50C4
type: fact
---

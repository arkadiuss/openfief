---
title: Core Only With Instance Token = true
status: active
text_ref: design-decisions:D-59#1
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-59-1
  - lane:strict
  - fact:property_value
fact_kind: property_value
subject_key: plugins.core_to_plugin_calls
property_key: core_only_with_instance_token
canonical_key: docs-design-md:plugins.core_to_plugin_calls:core_only_with_instance_token:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-CF03A08EF5A303C7
claim_text: 'A plugin must accept calls only from the core: its NetworkPolicy must admit ingress from the core alone, and each call must carry a short-lived ServiceAccount token whose audience is that plugin instance.'
origin:
  kind: agent
  recorded_at: '2026-10-28T06:09:38.571Z'
id: FACT-PROP-CF4B222B4E769C4A
type: fact
---

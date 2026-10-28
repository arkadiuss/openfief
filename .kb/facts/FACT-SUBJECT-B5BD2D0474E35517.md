---
title: Plugins Core To Plugin Calls
status: active
text_ref: design-decisions:D-59#1
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-59-1
  - lane:strict
  - fact:subject
  - vocabulary:new-subject
fact_kind: subject
subject_key: plugins.core_to_plugin_calls
canonical_key: plugins.core_to_plugin_calls
claim_key: CLAIM-CF03A08EF5A303C7
claim_text: 'A plugin must accept calls only from the core: its NetworkPolicy must admit ingress from the core alone, and each call must carry a short-lived ServiceAccount token whose audience is that plugin instance.'
origin:
  kind: agent
  recorded_at: '2026-10-28T06:09:37.353Z'
id: FACT-SUBJECT-B5BD2D0474E35517
type: fact
---

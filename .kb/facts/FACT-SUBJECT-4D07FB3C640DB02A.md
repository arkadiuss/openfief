---
title: E2e Core Plugin Calls
status: active
text_ref: docs/DESIGN.md#D-200
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-200
  - lane:strict
  - fact:subject
  - vocabulary:new-subject
fact_kind: subject
subject_key: e2e.core_plugin_calls
canonical_key: e2e.core_plugin_calls
claim_key: CLAIM-B7B0E0D9879D0282
claim_text: On both kind and k3d, a call to a connector from any pod other than the core must fail, a call from the core without the core's valid token for that instance must be refused, and the core's plugin port must refuse a plugin's call without a valid view plugin token.
origin:
  kind: agent
  recorded_at: '2026-11-06T20:50:04.586Z'
id: FACT-SUBJECT-4D07FB3C640DB02A
type: fact
---

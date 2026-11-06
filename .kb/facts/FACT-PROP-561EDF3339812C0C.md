---
title: One Item Per Failing Instance Run = true
status: active
text_ref: docs/DESIGN.md#D-196
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-196
  - lane:strict
  - fact:property_value
fact_kind: property_value
subject_key: attention.failing_sync_check
property_key: one_item_per_failing_instance_run
canonical_key: mcp-kibi-model-requirement:attention.failing_sync_check:one_item_per_failing_instance_run:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-ECCBD4961D87307A
claim_text: The attention sweep must raise one item of kind connector-sync-failing per connector instance whose sync keeps failing, with the instance as subject and when the failures began as detail, so it stays acknowledged while the same failures go on, resolves after the next success and opens again for a new run of failures.
origin:
  kind: agent
  recorded_at: '2026-11-06T01:26:54.061Z'
id: FACT-PROP-561EDF3339812C0C
type: fact
---

---
title: Waits With Bounded Backoff = true
status: active
text_ref: docs/dev/demo.md#demo-data
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-dev-demo-md-demo-data
  - lane:strict
  - fact:property_value
fact_kind: property_value
subject_key: demo.seed_database_wait
property_key: waits_with_bounded_backoff
canonical_key: mcp-kibi-model-requirement:demo.seed_database_wait:waits_with_bounded_backoff:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-CD04E6DB03EC835A
claim_text: The demo loader must wait for its database to accept connections, retrying with a capped backoff for a bounded time, before it starts loading, and fail with the last error when the database still refuses at that bound.
origin:
  kind: agent
  recorded_at: '2026-11-03T17:15:14.012Z'
id: FACT-PROP-AEEA64060ED519B9
type: fact
---

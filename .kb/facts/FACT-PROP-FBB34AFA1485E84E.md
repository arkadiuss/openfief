---
title: Syncs When Due = true
status: active
text_ref: docs/DESIGN.md#D-49
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-49
  - lane:strict
  - fact:property_value
fact_kind: property_value
subject_key: ingest.scheduler
property_key: syncs_when_due
canonical_key: mcp-kibi-model-requirement:ingest.scheduler:syncs_when_due:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-ED7D7D8E5EF86F7D
claim_text: The core must sync every running connector instance when its sync policy says it is due, at once when it never synced and then every interval plus a random jitter after a success, in an opening batch until one completes and in transactions batches after it, and must not sync an instance that no longer runs while keeping its state.
origin:
  kind: agent
  recorded_at: '2026-10-28T16:50:29.576Z'
id: FACT-PROP-FBB34AFA1485E84E
type: fact
---

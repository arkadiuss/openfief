---
title: Syncs Due Instances = true
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
property_key: syncs_due_instances
canonical_key: mcp-kibi-model-requirement:ingest.scheduler:syncs_due_instances:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-7239C60159E2987E
claim_text: The core must sync every running connector instance when its sync policy says it is due, at once when it never synced and then every interval plus a random jitter after a success, and must not sync an instance that no longer runs while keeping its state.
origin:
  kind: agent
  recorded_at: '2026-10-29T17:20:48.171Z'
id: FACT-PROP-B0460D3EED7103D0
type: fact
---

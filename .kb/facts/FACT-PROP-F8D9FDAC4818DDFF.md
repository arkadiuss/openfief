---
title: One Sync Worker Unless Raised With Memory Limit = true
status: active
text_ref: design-decisions:D-189#1
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-189-1
  - lane:strict
  - fact:property_value
fact_kind: property_value
subject_key: core.connector_syncs
property_key: one_sync_worker_unless_raised_with_memory_limit
canonical_key: mcp-kibi-model-requirement:core.connector_syncs:one_sync_worker_unless_raised_with_memory_limit:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-0DBF91F0ABCFA082
claim_text: The core must run one connector sync worker unless the owner raises the worker count together with the core's memory limit.
origin:
  kind: agent
  recorded_at: '2026-11-03T18:24:10.228Z'
id: FACT-PROP-F8D9FDAC4818DDFF
type: fact
---

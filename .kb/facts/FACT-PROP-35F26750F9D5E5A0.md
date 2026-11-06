---
title: Owner Retries Sync Once = true
status: active
text_ref: design-decisions:D-196
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-196
  - lane:strict
  - fact:property_value
fact_kind: property_value
subject_key: ingest.sync_retry
property_key: owner_retries_sync_once
canonical_key: mcp-kibi-model-requirement:ingest.sync_retry:owner_retries_sync_once:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-F697BAF8E0B32439
claim_text: The core must let the owner start one sync of a connector instance at once through an owner-only service, after the authorizer allows a change of every account bound to it, and must refuse it with that reason while a sync of the instance runs.
origin:
  kind: agent
  recorded_at: '2026-11-06T04:23:00.995Z'
id: FACT-PROP-35F26750F9D5E5A0
type: fact
---

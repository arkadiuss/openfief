---
title: Retry Rate Limited Per Instance = true
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
property_key: retry_rate_limited_per_instance
canonical_key: mcp-kibi-model-requirement:ingest.sync_retry:retry_rate_limited_per_instance:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-CDE2FFDE3B9D78D0
claim_text: The core must refuse a retry of a connector instance's sync less than a minute after the last one it started.
origin:
  kind: agent
  recorded_at: '2026-11-06T04:23:04.830Z'
id: FACT-PROP-5FD72C7639C7DAB6
type: fact
---

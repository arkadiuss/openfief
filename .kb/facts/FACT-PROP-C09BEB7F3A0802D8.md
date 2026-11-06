---
title: Keeps Failing After Three Over 15 Minutes = true
status: active
text_ref: design-decisions:D-198#1
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-198-1
  - lane:strict
  - fact:property_value
fact_kind: property_value
subject_key: ingest.sync_status
property_key: keeps_failing_after_three_over_15_minutes
canonical_key: mcp-kibi-model-requirement:ingest.sync_status:keeps_failing_after_three_over_15_minutes:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-FC88C5C98AF98965
claim_text: A connector's sync must keep failing after three failed attempts in a row since its last success, spanning at least 15 minutes, and one success must clear it.
origin:
  kind: agent
  recorded_at: '2026-11-06T06:15:22.326Z'
id: FACT-PROP-C09BEB7F3A0802D8
type: fact
---

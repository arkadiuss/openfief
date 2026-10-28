---
title: Single Flight = true
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
property_key: single_flight
canonical_key: mcp-kibi-model-requirement:ingest.scheduler:single_flight:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-950D1300614267D2
claim_text: At most one sync of a connector instance must run at a time, within one core process and across core replicas through a stored claim that expires, and each attempt must be bounded by a timeout whose expiry counts as a failed attempt.
origin:
  kind: agent
  recorded_at: '2026-10-28T16:50:39.534Z'
id: FACT-PROP-8ECD47DA61425FDE
type: fact
---

---
title: Checks Version Before Writes = true
status: active
text_ref: design-decisions:D-188#1
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-188-1
  - lane:strict
  - fact:property_value
fact_kind: property_value
subject_key: views.ghostfolio_sync
property_key: checks_version_before_writes
canonical_key: mcp-kibi-model-requirement:views.ghostfolio_sync:checks_version_before_writes:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-45C8546E8F31333D
claim_text: The Ghostfolio plugin must check that the Ghostfolio it talks to is a version it supports before it changes or deletes anything there.
origin:
  kind: agent
  recorded_at: '2026-11-03T14:31:35.042Z'
id: FACT-PROP-6AF6CBA755DC6BA3
type: fact
---

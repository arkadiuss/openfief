---
title: Opening Upload Starts = true
status: active
text_ref: docs/DESIGN.md#D-160
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-160
  - lane:strict
  - fact:property_value
fact_kind: property_value
subject_key: ingest.account_start
property_key: opening_upload_starts
canonical_key: mcp-kibi-model-requirement:ingest.account_start:opening_upload_starts:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-D24A33B5E8134AAC
claim_text: An opening upload must record its account's go-live at once, even when nothing can be booked yet, so later uploads are transactions and waiting positions join the opening on that date.
origin:
  kind: agent
  recorded_at: '2026-11-01T16:50:53.857Z'
id: FACT-PROP-F345DA41FCCD767E
type: fact
---

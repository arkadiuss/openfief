---
title: Records Every Attempt With Class = true
status: active
text_ref: docs/DESIGN.md#D-196
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-196
  - lane:strict
  - fact:property_value
fact_kind: property_value
subject_key: ingest.sync_history
property_key: records_every_attempt_with_class
canonical_key: mcp-kibi-model-requirement:ingest.sync_history:records_every_attempt_with_class:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-0B1092DF451DC233
claim_text: Every connector sync attempt must be recorded in an append-only history with its batch, mode, start, end and next sync, and a failed one with how many attempts in a row have failed and an error class taken from the known errors in its chain, never the error's text.
origin:
  kind: agent
  recorded_at: '2026-11-05T23:39:05.471Z'
id: FACT-PROP-097EDC4DEDF9DAC0
type: fact
---

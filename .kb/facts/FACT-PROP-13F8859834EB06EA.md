---
title: Keeps Failing After Three Over Fifteen Minutes = true
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
subject_key: ingest.sync_status
property_key: keeps_failing_after_three_over_fifteen_minutes
canonical_key: mcp-kibi-model-requirement:ingest.sync_status:keeps_failing_after_three_over_fifteen_minutes:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-32DF80F61CB1CFF4
claim_text: A connector instance's sync must count as one that keeps failing when at least three attempts in a row have failed since its latest success and they span at least fifteen minutes from the first one's start to the latest one's end, and one success must clear it.
origin:
  kind: agent
  recorded_at: '2026-11-05T23:39:10.622Z'
id: FACT-PROP-13F8859834EB06EA
type: fact
---

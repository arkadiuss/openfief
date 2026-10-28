---
title: Validated Before Storing = true
status: active
text_ref: docs/DESIGN.md#D-36
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-36
  - lane:strict
  - fact:property_value
fact_kind: property_value
subject_key: ingest.index_reading
property_key: validated_before_storing
canonical_key: mcp-kibi-model-requirement:ingest.index_reading:validated_before_storing:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-B131C15AFB7638FD
claim_text: An index reading must name CPI or the NBP reference rate, a date and a source, and a CPI reading must be dated on the first of the month it measures; when any reading in a delivery is invalid, the core must fail the batch and import none of it.
origin:
  kind: agent
  recorded_at: '2026-10-28T08:56:27.796Z'
id: FACT-PROP-E156396CB400D6B3
type: fact
---

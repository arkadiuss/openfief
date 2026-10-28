---
title: Reports Rows As Problems = true
status: active
text_ref: docs/DESIGN.md#D-7
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-7
  - lane:strict
  - fact:property_value
fact_kind: property_value
subject_key: ingest.mapping
property_key: reports_rows_as_problems
canonical_key: mcp-kibi-model-requirement:ingest.mapping:reports_rows_as_problems:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-C1B6E7573AB6C897
claim_text: A statement row that a connector cannot or will not book must become an ingest problem, unsupported or skipped, naming the row and the reason without amounts or account numbers, and must not fail the rest of the import batch.
origin:
  kind: agent
  recorded_at: '2026-10-28T06:24:16.746Z'
id: FACT-PROP-AF20A30D2BE75045
type: fact
---

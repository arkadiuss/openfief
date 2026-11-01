---
title: Count Unsupported Rows = true
status: active
text_ref: docs/DESIGN.md#D-130
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-130
  - lane:strict
  - fact:property_value
fact_kind: property_value
subject_key: ingest.import_batches
property_key: count_unsupported_rows
canonical_key: mcp-kibi-model-requirement:ingest.import_batches:count_unsupported_rows:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-8DF65D4E28A65188
claim_text: Each completed import batch must store how many of its rows OpenFief can't book yet, apart from rows left out on purpose.
origin:
  kind: agent
  recorded_at: '2026-10-31T18:32:51.106Z'
id: FACT-PROP-243198CBD872C97E
type: fact
---

---
title: Stores Delivered Index Levels = true
status: active
text_ref: docs/DESIGN.md#D-51
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-51
  - lane:strict
  - fact:property_value
fact_kind: property_value
subject_key: ingest.index_levels
property_key: stores_delivered_index_levels
canonical_key: mcp-kibi-model-requirement:ingest.index_levels:stores_delivered_index_levels:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-BBDB965E01F00324
claim_text: The core must translate a connector delivery's index levels into index levels keyed by the index name the connector sends, refuse one without a name, a date or a positive level, and store them with the connector instance as their source.
origin:
  kind: agent
  recorded_at: '2026-10-28T22:42:33.936Z'
id: FACT-PROP-8E3F40C770CC1478
type: fact
---

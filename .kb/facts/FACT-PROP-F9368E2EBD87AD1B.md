---
title: Stored Only From Allowed Instance With Source = true
status: active
text_ref: docs/DESIGN.md#D-179
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-179
  - lane:strict
  - fact:property_value
fact_kind: property_value
subject_key: ingest.index_levels
property_key: stored_only_from_allowed_instance_with_source
canonical_key: mcp-kibi-model-requirement:ingest.index_levels:stored_only_from_allowed_instance_with_source:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-E7D3A6DFC45B2F17
claim_text: The core must translate a connector delivery's index levels into index levels keyed by the index name the connector sends, refuse one without a name, a date or a positive level, and store those of an instance that may deliver them with the instance as their source.
origin:
  kind: agent
  recorded_at: '2026-11-02T22:25:08.870Z'
id: FACT-PROP-F9368E2EBD87AD1B
type: fact
---

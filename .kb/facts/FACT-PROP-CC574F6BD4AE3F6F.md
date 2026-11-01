---
title: Starts On Charts Start = true
status: active
text_ref: docs/DESIGN.md#D-138
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-138
  - lane:strict
  - fact:property_value
fact_kind: property_value
subject_key: views.performance_query
property_key: starts_on_charts_start
canonical_key: mcp-kibi-model-requirement:views.performance_query:starts_on_charts_start:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-735604A2D566629B
claim_text: The performance query must start on the start the caller chose, by default on the charts' start, the latest day an account was seeded from a starting balance or else the first transaction's date, a chosen day before the first transaction on that day, end by default at the latest valued-through date of any account, name the default start, and report whether any revaluation is pending or running.
origin:
  kind: agent
  recorded_at: '2026-11-01T13:01:39.013Z'
id: FACT-PROP-CC574F6BD4AE3F6F
type: fact
---

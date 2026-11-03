---
title: Serialized Across Processes = true
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
subject_key: valuation.river_migrations
property_key: serialized_across_processes
canonical_key: mcp-kibi-model-requirement:valuation.river_migrations:serialized_across_processes:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-BA71BFF910DC244F
claim_text: River's migrations must run under an advisory lock shared by every process that migrates the database, so that concurrent migrations, such as the core's and the demo loader's, all succeed and apply each version once.
origin:
  kind: agent
  recorded_at: '2026-11-03T17:15:04.835Z'
id: FACT-PROP-8FCA59C1D42231E5
type: fact
---

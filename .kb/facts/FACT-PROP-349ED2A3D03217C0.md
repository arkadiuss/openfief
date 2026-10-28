---
title: Refuses Unmarked Database = true
status: active
text_ref: docs/DESIGN.md#GR-12
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-gr-12
  - lane:strict
  - fact:property_value
fact_kind: property_value
subject_key: demo.loader
property_key: refuses_unmarked_database
canonical_key: mcp-kibi-model-requirement:demo.loader:refuses_unmarked_database:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-276A49FAB9D7BECB
claim_text: The demo loader must refuse, before writing anything, a database that holds ledger or valuation data and has no demo marker, with or without --reset-demo; it may load an empty database, which it first marks as a demo database, and may reset only a database that carries the marker.
origin:
  kind: agent
  recorded_at: '2026-10-28T09:31:01.649Z'
id: FACT-PROP-349ED2A3D03217C0
type: fact
---

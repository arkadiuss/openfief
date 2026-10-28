---
title: Imports Through Connector = true
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
subject_key: ingest.ingest_path
property_key: imports_through_connector
canonical_key: mcp-kibi-model-requirement:ingest.ingest_path:imports_through_connector:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-4007313BD110B132
claim_text: A connector's ingest path, from a statement file or a pull, must map the statement with that institution's connector and import the resulting transactions and balance assertions into the ledger as one import batch, and must import nothing when the institution has no connector, the statement cannot be read or the pull fails.
origin:
  kind: agent
  recorded_at: '2026-10-28T06:37:26.816Z'
id: FACT-PROP-AC9E68FB7221D8B7
type: fact
---

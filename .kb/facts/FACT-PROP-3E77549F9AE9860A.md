---
title: Imported By The Core = true
status: active
text_ref: docs/DESIGN.md#D-49
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-49
  - lane:strict
  - fact:property_value
fact_kind: property_value
subject_key: ingest.connector_batches
property_key: imported_by_the_core
canonical_key: mcp-kibi-model-requirement:ingest.connector_batches:imported_by_the_core:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-FC88B8A6209245E9
claim_text: The core must call a connector instance's Sync or ImportFile with a batch ID its caller supplies and reuses on a retry, import the delivery as one import batch with source sync or file, store its prices, FX rates and index readings with the instance as source, return a sync's next watermark, and import nothing when the connector fails or the delivery is invalid.
origin:
  kind: agent
  recorded_at: '2026-10-28T08:55:13.363Z'
id: FACT-PROP-3E77549F9AE9860A
type: fact
---

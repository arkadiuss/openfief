---
title: Ingest Connector Batches
status: active
text_ref: docs/DESIGN.md#D-49
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-49
  - lane:strict
  - fact:subject
  - vocabulary:new-subject
fact_kind: subject
subject_key: ingest.connector_batches
canonical_key: ingest.connector_batches
claim_key: CLAIM-FC88B8A6209245E9
claim_text: The core must call a connector instance's Sync or ImportFile with a batch ID its caller supplies and reuses on a retry, import the delivery as one import batch with source sync or file, store its prices, FX rates and index readings with the instance as source, return a sync's next watermark, and import nothing when the connector fails or the delivery is invalid.
origin:
  kind: agent
  recorded_at: '2026-10-28T08:55:08.703Z'
id: FACT-SUBJECT-26599C6DFD7167ED
type: fact
---

---
title: Retries Same Batch = true
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
subject_key: ingest.scheduler
property_key: retries_same_batch
canonical_key: mcp-kibi-model-requirement:ingest.scheduler:retries_same_batch:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-FA06080FF364669F
claim_text: A failed sync attempt must keep its import batch open and be retried with the same batch ID, mode and watermark after a backoff that starts at the first retry delay and doubles with each consecutive failure up to the maximum, and the watermark must be stored only when the batch completes, after both the ledger import and the valuation inputs succeeded.
origin:
  kind: agent
  recorded_at: '2026-10-28T16:50:34.512Z'
id: FACT-PROP-8707CE1DB7B8C92A
type: fact
---

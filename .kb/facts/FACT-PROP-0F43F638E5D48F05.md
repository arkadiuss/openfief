---
title: Carries Bound Holdings = true
status: active
text_ref: docs/DESIGN.md#D-117
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-117
  - lane:strict
  - fact:property_value
fact_kind: property_value
subject_key: ingest.sync_request
property_key: carries_bound_holdings
canonical_key: mcp-kibi-model-requirement:ingest.sync_request:carries_bound_holdings:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-4512B4382550D09D
claim_text: Each sync request to a connector instance bound to accounts must carry, for each bound account once in ID order, its go-live date, tax wrapper, non-zero holdings and open lots today, and the stored series terms of its treasury bonds; an instance bound to no account gets none.
origin:
  kind: agent
  recorded_at: '2026-10-31T13:25:43.257Z'
id: FACT-PROP-0F43F638E5D48F05
type: fact
---

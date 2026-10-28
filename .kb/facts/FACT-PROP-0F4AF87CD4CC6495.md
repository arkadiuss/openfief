---
title: Scoped To Instance = true
status: active
text_ref: docs/DESIGN.md#D-48
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-48
  - lane:strict
  - fact:property_value
fact_kind: property_value
subject_key: ingest.connector_transaction_keys
property_key: scoped_to_instance
canonical_key: mcp-kibi-model-requirement:ingest.connector_transaction_keys:scoped_to_instance:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-AF3E361081306D21
claim_text: The core must prefix the key source of every transaction a connector instance delivers with plugin:<instance>:, so no connector can write or supersede another instance's or an in-core connector's transactions.
origin:
  kind: agent
  recorded_at: '2026-10-28T08:55:35.635Z'
id: FACT-PROP-0F4AF87CD4CC6495
type: fact
---

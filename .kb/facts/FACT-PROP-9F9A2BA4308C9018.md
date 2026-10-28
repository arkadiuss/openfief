---
title: Exact Messages With Batch Id = true
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
subject_key: connector.contract
property_key: exact_messages_with_batch_id
canonical_key: mcp-kibi-model-requirement:connector.contract:exact_messages_with_batch_id:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-ED878374C3A9623B
claim_text: The connector contract, openfief.connector.v1.ConnectorService, must carry amounts, prices, rates and quantities as exact decimal strings, dates as YYYY-MM-DD strings and currencies as ISO 4217 codes, and every Sync or ImportFile call must carry a batch ID the core generates, which the connector's delivery echoes.
origin:
  kind: agent
  recorded_at: '2026-10-28T08:01:07.251Z'
id: FACT-PROP-9F9A2BA4308C9018
type: fact
---

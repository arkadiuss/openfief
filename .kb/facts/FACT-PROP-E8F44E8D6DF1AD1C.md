---
title: Connector Eodhd Api Token Role = true
status: active
text_ref: docs/DESIGN.md#D-89
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-89
  - lane:strict
  - fact:property_value
fact_kind: property_value
subject_key: plugins.eodhd_config
property_key: connector_eodhd_api_token_role
canonical_key: mcp-kibi-model-requirement:plugins.eodhd_config:connector_eodhd_api_token_role:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-EA6ACC3BB2F37708
claim_text: The EODHD plugin's manifest must ask only for the connector capability, egress to eodhd.com and its api-token Secret holding the owner's API key, and the plugin must refuse a config that names an invalid ISIN, an unsupported exchange, or an instrument or index twice.
origin:
  kind: agent
  recorded_at: '2026-10-29T18:02:26.643Z'
id: FACT-PROP-E8F44E8D6DF1AD1C
type: fact
---

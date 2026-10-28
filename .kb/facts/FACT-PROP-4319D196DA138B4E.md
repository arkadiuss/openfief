---
title: Asks Only For Eodhd And Its Key = true
status: active
text_ref: docs/DESIGN.md#D-22
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-22
  - lane:strict
  - fact:property_value
fact_kind: property_value
subject_key: plugins.eodhd_connector
property_key: asks_only_for_eodhd_and_its_key
canonical_key: mcp-kibi-model-requirement:plugins.eodhd_connector:asks_only_for_eodhd_and_its_key:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-3AF714C20F8C39A4
claim_text: The EODHD plugin's manifest must ask only for the connector capability, egress to eodhd.com and the eodhd-credentials Secret holding the owner's API key, and the plugin must refuse a config that names an invalid ISIN, an unsupported exchange, or an instrument or index twice.
origin:
  kind: agent
  recorded_at: '2026-10-28T21:37:47.550Z'
id: FACT-PROP-4319D196DA138B4E
type: fact
---

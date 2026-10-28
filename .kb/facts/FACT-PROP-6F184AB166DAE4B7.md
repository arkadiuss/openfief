---
title: Syncs Per Series Watermark = true
status: active
text_ref: docs/DESIGN.md#D-51
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-51
  - lane:strict
  - fact:property_value
fact_kind: property_value
subject_key: plugins.eodhd_connector
property_key: syncs_per_series_watermark
canonical_key: mcp-kibi-model-requirement:plugins.eodhd_connector:syncs_per_series_watermark:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-A90B0A001F621EC3
claim_text: The EODHD plugin must answer Sync with the end-of-day closes of each configured instrument and the levels of each configured market index after that series' own watermark mark, or from its configured start, up to today, from institution eodhd; must report an unlisted, ambiguous or unknown series as an ingest problem and still deliver the rest; must answer Unavailable when EODHD fails and FailedPrecondition without an API key; and must refuse ImportFile.
origin:
  kind: agent
  recorded_at: '2026-10-28T21:37:42.137Z'
id: FACT-PROP-6F184AB166DAE4B7
type: fact
---

---
title: Kept Per Isin Append Only = true
status: active
text_ref: docs/DESIGN.md#D-197
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-197
  - lane:strict
  - fact:property_value
fact_kind: property_value
subject_key: ledger.instrument_names
property_key: kept_per_isin_append_only
canonical_key: mcp-kibi-model-requirement:ledger.instrument_names:kept_per_isin_append_only:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-00543E24C6B0D741
claim_text: 'The ledger must keep instruments'' display names per ISIN, append-only: the first name a connector delivers stays, the owner''s latest rename wins, names are cleaned and capped, and only the owner may rename, allowed to change every account that posts to the instrument.'
origin:
  kind: agent
  recorded_at: '2026-11-06T06:59:14.474Z'
id: FACT-PROP-2C43B38DA0F06CB6
type: fact
---

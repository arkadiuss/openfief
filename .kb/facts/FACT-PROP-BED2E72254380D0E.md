---
title: Carries Payout Setting = true
status: active
text_ref: docs/DESIGN.md#D-135
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-135
  - lane:strict
  - fact:property_value
fact_kind: property_value
subject_key: ingest.sync_request
property_key: carries_payout_setting
canonical_key: mcp-kibi-model-requirement:ingest.sync_request:carries_payout_setting:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-E25C6ABB0EFC4888
claim_text: Each sync request must carry each bound account's payout setting and the day it took effect, and a plugin must refuse an unknown setting or a paid-out one without a valid day.
origin:
  kind: agent
  recorded_at: '2026-11-01T07:45:13.060Z'
id: FACT-PROP-BED2E72254380D0E
type: fact
---

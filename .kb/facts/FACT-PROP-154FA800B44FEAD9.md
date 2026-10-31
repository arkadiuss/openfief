---
title: Deletes Unused Account = true
status: active
text_ref: docs/DESIGN.md#D-120
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-120
  - lane:strict
  - fact:property_value
fact_kind: property_value
subject_key: ingest.account_service
property_key: deletes_unused_account
canonical_key: mcp-kibi-model-requirement:ingest.account_service:deletes_unused_account:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-241F650A010D5FC0
claim_text: The core must delete an account that holds no transactions or balance assertions, with its bindings, broker account mappings, tax wrapper, fee rule and stored valuations, for the owner only, and refuse one that holds any without changing anything.
origin:
  kind: agent
  recorded_at: '2026-10-31T14:35:19.202Z'
id: FACT-PROP-154FA800B44FEAD9
type: fact
---

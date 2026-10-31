---
title: Remaps Unused Broker Account = true
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
property_key: remaps_unused_broker_account
canonical_key: mcp-kibi-model-requirement:ingest.account_service:remaps_unused_broker_account:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-A9C4D60E1D91016C
claim_text: The core must remap a broker account to another account or to none, for the owner only, while the account it is mapped to holds no transaction from its connector instance, and refuse it otherwise.
origin:
  kind: agent
  recorded_at: '2026-10-31T14:35:23.790Z'
id: FACT-PROP-FACA628A378E0174
type: fact
---

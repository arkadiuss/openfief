---
title: Seeded After Mapping = true
status: active
text_ref: docs/DESIGN.md#D-23
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-23
  - lane:strict
  - fact:property_value
fact_kind: property_value
subject_key: ingest.broker_accounts
property_key: seeded_after_mapping
canonical_key: mcp-kibi-model-requirement:ingest.broker_accounts:seeded_after_mapping:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-6BCEA07426B522FA
claim_text: A broker account mapped after its connector instance was seeded must make the instance's next sync an opening batch, which seeds it, and an opening batch must leave out the seeding transactions of a broker account already seeded while keeping its balance assertions.
origin:
  kind: agent
  recorded_at: '2026-10-29T11:36:58.255Z'
id: FACT-PROP-C1DD849DC6CE72C1
type: fact
---

---
title: Serializes Version Appends = true
status: active
text_ref: docs/DESIGN.md#D-18
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-18
  - lane:strict
  - fact:property_value
fact_kind: property_value
subject_key: ledger.import
property_key: serializes_version_appends
canonical_key: mcp-kibi-model-requirement:ledger.import:serializes_version_appends:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-7F98397E81D0B356
claim_text: 'Two imports racing for the same next version of a key must never both append it: the store must refuse the second append with a version conflict, and the import must reload and decide again, failing with a version conflict error after a bounded number of attempts.'
origin:
  kind: agent
  recorded_at: '2026-10-28T06:05:41.015Z'
id: FACT-PROP-0F4FC59DD46F18C8
type: fact
---

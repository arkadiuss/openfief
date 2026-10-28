---
title: Refuses Reseeding = true
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
subject_key: ledger.accounts
property_key: refuses_reseeding
canonical_key: mcp-kibi-model-requirement:ledger.accounts:refuses_reseeding:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-E7DBDB353368E19A
claim_text: Seeding an account whose current transactions already include an opening lot or opening cash must be refused with a typed already-seeded error, unless every incoming transaction of the account equals its current version, in which case the seeding changes nothing and reports the account as already seeded.
origin:
  kind: agent
  recorded_at: '2026-10-28T08:39:13.591Z'
id: FACT-PROP-D33FB3FF710FBCB4
type: fact
---

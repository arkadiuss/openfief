---
title: Seeds Atomically = true
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
property_key: seeds_atomically
canonical_key: mcp-kibi-model-requirement:ledger.accounts:seeds_atomically:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-F5D68475971E0AB7
claim_text: The already-seeded check and the seeding's appends must run in one database transaction that locks each seeded account first, so that two concurrent seedings of one account cannot both succeed.
origin:
  kind: agent
  recorded_at: '2026-10-28T08:39:21.927Z'
id: FACT-PROP-E6C58A08AC6A8B16
type: fact
---

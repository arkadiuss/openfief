---
title: New Versions Same Keys = true
status: active
text_ref: docs/DESIGN.md#D-214
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-214
  - lane:strict
  - fact:property_value
fact_kind: property_value
subject_key: ledger.rekey
property_key: new_versions_same_keys
canonical_key: mcp-kibi-model-requirement:ledger.rekey:new_versions_same_keys:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-5730629E84BFDF73
claim_text: The ledger must re-key an instrument by appending, in one database transaction, a new version under the same key of each current transaction that posts to it, with only its holding postings moved, running the import hooks, carrying its display names over, and changing nothing when run again.
origin:
  kind: agent
  recorded_at: '2026-11-07T22:45:20.863Z'
id: FACT-PROP-6FDE7659F1CE8F9B
type: fact
---

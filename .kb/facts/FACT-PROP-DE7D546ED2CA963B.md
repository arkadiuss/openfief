---
title: Supersedes Without Erasing = true
status: active
text_ref: docs/DESIGN.md#D-100
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-100
  - lane:strict
  - fact:property_value
fact_kind: property_value
subject_key: ledger.deletion
property_key: supersedes_without_erasing
canonical_key: mcp-kibi-model-requirement:ledger.deletion:supersedes_without_erasing:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-C465E8F1865B2A99
claim_text: A deletion must be a new version of its transaction's key with no postings that supersedes the current version, keeps every earlier version, leaves the transaction out of current transactions and everything derived from them, and must be refused for a key that has no version.
origin:
  kind: agent
  recorded_at: '2026-10-30T05:01:50.039Z'
id: FACT-PROP-DE7D546ED2CA963B
type: fact
---

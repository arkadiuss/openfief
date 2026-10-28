---
title: Flagged On Import = true
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
subject_key: ledger.balance_assertion
property_key: flagged_on_import
canonical_key: mcp-kibi-model-requirement:ledger.balance_assertion:flagged_on_import:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-1FC04A0118B07C28
claim_text: Importing must check every balance assertion of each account it touches after appending, store each result as matched or as flagged with the derived balance, and keep the import when an assertion is flagged.
origin:
  kind: agent
  recorded_at: '2026-10-28T06:05:46.065Z'
id: FACT-PROP-998BF6D15BC95055
type: fact
---

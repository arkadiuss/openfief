---
title: Appends Decided Versions Only = true
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
property_key: appends_decided_versions_only
canonical_key: mcp-kibi-model-requirement:ledger.import:appends_decided_versions_only:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-B04DE71EC7253A00
claim_text: Importing must decide each incoming transaction against every stored version of its key, append a new key as version 1 and changed content as the next version, append nothing for unchanged content, and never change or remove a stored version.
origin:
  kind: agent
  recorded_at: '2026-10-28T06:05:35.284Z'
id: FACT-PROP-23FCF0713D4F5EC5
type: fact
---

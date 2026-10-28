---
title: Append Only Enforced = true
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
subject_key: ledger.store
property_key: append_only_enforced
canonical_key: mcp-kibi-model-requirement:ledger.store:append_only_enforced:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-29A26B461CA33E84
claim_text: The database must refuse any update, delete or truncate of stored transaction versions and their postings, any posting added to a stored version, and any version that is not the next of its key, for every role, and loading must detect a stored version whose content no longer matches its fingerprint.
origin:
  kind: agent
  recorded_at: '2026-10-28T06:29:40.292Z'
id: FACT-PROP-3EDBBEB7DDCD756D
type: fact
---

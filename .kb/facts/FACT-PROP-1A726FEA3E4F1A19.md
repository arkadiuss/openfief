---
title: Read Changed Or Rotated = false
status: active
text_ref: mcp://kibi/model-requirement
tags:
  - strict-modeling
  - confidence:0.80
  - confidence-band:medium
  - provenance:mcp-kibi-model-requirement
  - lane:strict
  - fact:property_value
fact_kind: property_value
subject_key: credentials_job.secrets
property_key: read_changed_or_rotated
canonical_key: mcp-kibi-model-requirement:credentials_job.secrets:read_changed_or_rotated:eq:false
operator: eq
value_type: bool
value_bool: false
claim_key: CLAIM-932B6DDB5213DA5E
claim_text: The Job never reads, changes or rotates a Secret, so an owner who feeds the platform credentials from their own tooling first (D-29) keeps theirs.
origin:
  kind: agent
  recorded_at: '2026-11-13T07:42:56.331Z'
id: FACT-PROP-1A726FEA3E4F1A19
type: fact
---

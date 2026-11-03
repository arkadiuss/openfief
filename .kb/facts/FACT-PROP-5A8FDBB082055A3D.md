---
title: Failures Logged Without Values = true
status: active
text_ref: docs/DESIGN.md#GR-12
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-gr-12
  - lane:strict
  - fact:property_value
fact_kind: property_value
subject_key: core.logging
property_key: failures_logged_without_values
canonical_key: mcp-kibi-model-requirement:core.logging:failures_logged_without_values:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-4B500FB65A6E517B
claim_text: Every context's Connect services, the built-in UI and the core's attention sweep must log a failure only by the known errors in its chain, or else by its innermost error's type, never with the amounts, holdings, ISINs or account numbers its text may carry.
origin:
  kind: agent
  recorded_at: '2026-11-03T03:24:24.941Z'
id: FACT-PROP-5A8FDBB082055A3D
type: fact
---

---
title: Keeps Previous Version Until Approved = true
status: active
text_ref: docs/DESIGN.md#D-16
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-16
  - lane:strict
  - fact:property_value
fact_kind: property_value
subject_key: plugins.rollout
property_key: keeps_previous_version_until_approved
canonical_key: mcp-kibi-model-requirement:plugins.rollout:keeps_previous_version_until_approved:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-E059A7B53B3729C9
claim_text: When the current version of a plugin requests permissions its grant does not approve, the previously deployed version must keep running under the current grant, and nothing must run when no version was deployed before.
origin:
  kind: agent
  recorded_at: '2026-10-28T06:36:27.565Z'
id: FACT-PROP-9E5B6606FB3A2994
type: fact
---

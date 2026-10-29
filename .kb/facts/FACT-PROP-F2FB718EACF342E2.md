---
title: Marked At Read Time = true
status: active
text_ref: docs/DESIGN.md#D-58
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-58
  - lane:strict
  - fact:property_value
fact_kind: property_value
subject_key: views.removed_instance
property_key: marked_at_read_time
canonical_key: mcp-kibi-model-requirement:views.removed_instance:marked_at_read_time:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-0D14417AFD55E0E4
claim_text: Where the read API lists a transaction, it must name the connector instance that brought it in and mark it as coming from a removed instance when no Plugin of that name is installed, deriving the mark when read and marking nothing while the installed plugins are unknown.
origin:
  kind: agent
  recorded_at: '2026-10-29T03:29:50.609Z'
id: FACT-PROP-F2FB718EACF342E2
type: fact
---

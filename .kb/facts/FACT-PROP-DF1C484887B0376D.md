---
title: Flags Ungranted Or Overlong Lines = true
status: active
text_ref: docs/DESIGN.md#D-200
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-200
  - lane:strict
  - fact:property_value
fact_kind: property_value
subject_key: plugins.egress_audit
property_key: flags_ungranted_or_overlong_lines
canonical_key: mcp-kibi-model-requirement:plugins.egress_audit:flags_ungranted_or_overlong_lines:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-21092876C272547E
claim_text: The egress audit must fail when the egress proxy allowed a request to a host the asking instance's grant does not approve, or when a decision line holds a field beyond who asked, where to, the decision and its timing, a requested host that is not a bare host and port, or no instance or decision; it must only read, and must report counts, instances and hosts but never a value that could hold a request's contents.
origin:
  kind: agent
  recorded_at: '2026-11-07T03:31:00.639Z'
id: FACT-PROP-DF1C484887B0376D
type: fact
---

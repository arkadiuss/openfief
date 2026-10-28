---
title: Reports Pending Approval = true
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
subject_key: operator.plugin_status
property_key: reports_pending_approval
canonical_key: mcp-kibi-model-requirement:operator.plugin_status:reports_pending_approval:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-5D427B563C4C0A91
claim_text: When a Plugin's spec requests permissions its PluginGrant does not approve, the operator must keep running the previously deployed version and report exactly the unapproved permissions in status.pendingApproval and an ApprovalRequired condition.
origin:
  kind: agent
  recorded_at: '2026-10-28T07:00:12.883Z'
id: FACT-PROP-5AFCC995D6DC9B8F
type: fact
---

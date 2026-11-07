---
title: Runs After Demo Deploy And Smoke = true
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
property_key: runs_after_demo_deploy_and_smoke
canonical_key: mcp-kibi-model-requirement:plugins.egress_audit:runs_after_demo_deploy_and_smoke:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-179A7627DD27815A
claim_text: 'After each demo deploy and after the release smoke test''s install, the egress proxy''s log must be audited against the cluster''s PluginGrants: on the demo by a Job at the egress proxy''s verified digest, re-created by every deploy, with a Role that only lists pods and grants and reads pod logs.'
origin:
  kind: agent
  recorded_at: '2026-11-07T03:31:05.496Z'
id: FACT-PROP-4081481CFB2E01BB
type: fact
---

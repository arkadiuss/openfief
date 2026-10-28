---
title: Operator Plugin Reconcile
status: active
text_ref: docs/DESIGN.md#D-16
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-16
  - lane:strict
  - fact:subject
  - vocabulary:new-subject
fact_kind: subject
subject_key: operator.plugin_reconcile
canonical_key: operator.plugin_reconcile
claim_key: CLAIM-E2F84D82D3E62780
claim_text: The operator must run a Plugin only under the PluginGrant of the same name, and when nothing may run it must remove the plugin's Deployment, Service, config and ServiceAccount and make its NetworkPolicy deny everything.
origin:
  kind: agent
  recorded_at: '2026-10-28T07:00:03.350Z'
id: FACT-SUBJECT-1E06339900ED6C5A
type: fact
---

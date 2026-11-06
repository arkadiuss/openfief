---
title: E2e Plugin Isolation
status: active
text_ref: docs/DESIGN.md#D-200
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-200
  - lane:strict
  - fact:subject
  - vocabulary:new-subject
fact_kind: subject
subject_key: e2e.plugin_isolation
canonical_key: e2e.plugin_isolation
claim_key: CLAIM-FD88A29F704445E2
claim_text: On both kind and k3d, each plugin must run in its own pod, the XTB parser must have no egress at all, a granted plugin's request through the egress proxy for a host outside its grant must be refused, and a direct connection or a DNS lookup from a plugin pod must fail.
origin:
  kind: agent
  recorded_at: '2026-11-06T20:49:58.024Z'
id: FACT-SUBJECT-E5B3988DBF72D3FD
type: fact
---

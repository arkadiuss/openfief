---
title: Plugins Workload
status: active
text_ref: docs/DESIGN.md#D-27
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-27
  - lane:strict
  - fact:subject
  - vocabulary:new-subject
fact_kind: subject
subject_key: plugins.workload
canonical_key: plugins.workload
claim_key: CLAIM-0997891548AF7E46
claim_text: A plugin's compiled workload must run in the openfief namespace as its own ServiceAccount with no Kubernetes API token and no role, carry a projected token with audience openfief-core, and mount read-only only its effective Secrets.
origin:
  kind: agent
  recorded_at: '2026-10-28T05:28:29.879Z'
id: FACT-SUBJECT-580F1F6B1556993B
type: fact
---

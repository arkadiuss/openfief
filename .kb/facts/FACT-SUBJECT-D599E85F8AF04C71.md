---
title: Core Connectors
status: active
text_ref: docs/DESIGN.md#D-49
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-49
  - lane:strict
  - fact:subject
  - vocabulary:new-subject
fact_kind: subject
subject_key: core.connectors
canonical_key: core.connectors
claim_key: CLAIM-D9766EA1CD6588CE
claim_text: In a Kubernetes cluster core serve must run the connector scheduler in the background when connectors are on, the default in a pod, without affecting readiness, and wait for it on shutdown; it must import deliveries through the ledger and record their prices, FX rates and index readings through valuation's input use cases; core connector bind must bind an instance to a stored account and core connector sync must sync one instance at once.
origin:
  kind: agent
  recorded_at: '2026-10-28T17:22:59.910Z'
id: FACT-SUBJECT-D599E85F8AF04C71
type: fact
---

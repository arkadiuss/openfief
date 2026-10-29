---
title: Ingest Connector Binding
status: active
text_ref: docs/DESIGN.md#D-48
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-48
  - lane:strict
  - fact:subject
  - vocabulary:new-subject
fact_kind: subject
subject_key: ingest.connector_binding
canonical_key: ingest.connector_binding
claim_key: CLAIM-11CFCF371D6CF259
claim_text: A connector instance's delivery must be refused whole, before anything is imported, when any of its transactions or balance assertions is in an account the instance has no connector binding to, and a connector binding must only be added for a stored account and a valid instance name.
origin:
  kind: agent
  recorded_at: '2026-10-28T16:50:19.918Z'
id: FACT-SUBJECT-675A558467AB9700
type: fact
---

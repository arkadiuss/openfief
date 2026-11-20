---
title: Ingest Deliveries
status: active
text_ref: docs/DESIGN.md#D-182
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-182
  - lane:strict
  - fact:subject
  - vocabulary:new-subject
fact_kind: subject
subject_key: ingest.deliveries
canonical_key: ingest.deliveries
claim_key: CLAIM-C0B7D02B561F51CD
claim_text: A connector's signed manifest must declare the institutions it books for, and the core must refuse, whole, a delivery that names an institution the running version's manifest doesn't declare, importing and storing nothing of it.
origin:
  kind: agent
  recorded_at: '2026-11-02T22:19:58.894Z'
id: FACT-SUBJECT-CAD68A36D51B31B5
type: fact
part_of:
  - ingest.delivery
---
The institutions a connector's signed manifest declares and the whole refusal of a delivery naming another. It is a part of ingest.delivery, which also covers the connector contract messages a delivery is written in.
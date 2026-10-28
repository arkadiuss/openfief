---
title: Ingest Connector Transaction Keys
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
subject_key: ingest.connector_transaction_keys
canonical_key: ingest.connector_transaction_keys
claim_key: CLAIM-AF3E361081306D21
claim_text: The core must prefix the key source of every transaction a connector instance delivers with plugin:<instance>:, so no connector can write or supersede another instance's or an in-core connector's transactions.
origin:
  kind: agent
  recorded_at: '2026-10-28T08:55:30.175Z'
id: FACT-SUBJECT-88F1C2EA762103A7
type: fact
---

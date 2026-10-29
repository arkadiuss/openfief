---
title: Ingest Connector Delivery
status: active
text_ref: docs/DESIGN.md#D-22
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-22
  - lane:strict
  - fact:subject
  - vocabulary:new-subject
fact_kind: subject
subject_key: ingest.connector_delivery
canonical_key: ingest.connector_delivery
claim_key: CLAIM-0D1270AC0DDD9ED6
claim_text: The core must translate a delivery's cross_check_prices with the same rules as prices, give each the connector instance as its source, and store them only as cross-check prices, never as prices; a delivery that carries some must fail when no cross-check store is wired rather than drop them.
origin:
  kind: agent
  recorded_at: '2026-10-29T06:24:41.095Z'
id: FACT-SUBJECT-4022BDCF8FBBCB20
type: fact
---

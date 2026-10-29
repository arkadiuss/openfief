---
title: Cross Check Prices Only From Declared Connectors = true
status: active
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-112-1
  - lane:strict
  - fact:property_value
text_ref: design-decisions:D-112#1
fact_kind: property_value
subject_key: ingest.connector_delivery
property_key: cross_check_prices_only_from_declared_connectors
operator: eq
value_type: bool
value_bool: true
canonical_key: docs-design-md:ingest.connector_delivery:cross_check_prices_only_from_declared_connectors:eq:true
claim_key: CLAIM-6B088C7AD2B37B58
claim_text: The core must accept cross-check prices only from a connector whose manifest declares that it sends them.
origin:
  kind: agent
  recorded_at: '2026-10-29T10:54:50.766Z'
id: FACT-PROP-0970C3AF0387F3E2
type: fact
---

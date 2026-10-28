---
title: Connector Contract
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
subject_key: connector.contract
canonical_key: connector.contract
claim_key: CLAIM-ED878374C3A9623B
claim_text: The connector contract, openfief.connector.v1.ConnectorService, must carry amounts, prices, rates and quantities as exact decimal strings, dates as YYYY-MM-DD strings and currencies as ISO 4217 codes, and every Sync or ImportFile call must carry a batch ID the core generates, which the connector's delivery echoes.
origin:
  kind: agent
  recorded_at: '2026-10-28T08:01:03.798Z'
id: FACT-SUBJECT-F2AEADC36D3854FB
type: fact
---

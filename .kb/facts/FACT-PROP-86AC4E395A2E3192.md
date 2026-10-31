---
title: List Only To Declared Granted Market Data Connectors = true
status: active
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-123-1
  - lane:strict
  - fact:property_value
text_ref: design-decisions:D-123#1
fact_kind: property_value
subject_key: ingest.market_data_sync
property_key: list_only_to_declared_granted_market_data_connectors
operator: eq
value_type: bool
value_bool: true
canonical_key: kb-requirements-req-ingest-market-data-list-declared-granted-md:ingest.market_data_sync:list_only_to_declared_granted_market_data_connectors:eq:true
claim_key: CLAIM-19E87FB1C8E875E6
claim_text: Only a connector whose manifest declares it a market-data connector, and whose grant approves that, must receive D-93's list of held instruments and model targets.
origin:
  kind: agent
  recorded_at: '2026-10-31T14:34:42.354Z'
id: FACT-PROP-86AC4E395A2E3192
type: fact
---

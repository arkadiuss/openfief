---
title: Reports Broker Accounts = true
status: active
text_ref: docs/DESIGN.md#D-72
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-72
  - lane:strict
  - fact:property_value
fact_kind: property_value
subject_key: ingest.ibkr_flex
property_key: reports_broker_accounts
canonical_key: mcp-kibi-model-requirement:ingest.ibkr_flex:reports_broker_accounts:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-E472BF36C257D7BB
claim_text: The IBKR Flex connector must take no account config, name every statement's entries by its IBKR account with keys whose source names no account, and report every IBKR account it saw, so that the core's mapping gives the same entries and keys as a configured account map.
origin:
  kind: agent
  recorded_at: '2026-10-29T11:37:12.238Z'
id: FACT-PROP-B82DFAE47CFF7418
type: fact
---

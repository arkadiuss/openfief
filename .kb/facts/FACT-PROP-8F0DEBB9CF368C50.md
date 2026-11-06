---
title: Serves Synthetic Institution Answers = true
status: active
text_ref: design-session:connector-endpoint-ca
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-session-connector-endpoint-ca
  - lane:strict
  - fact:property_value
fact_kind: property_value
subject_key: e2e.fixture_server
property_key: serves_synthetic_institution_answers
canonical_key: mcp-kibi-model-requirement:e2e.fixture_server:serves_synthetic_institution_answers:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-03537CCD6AEE3528
claim_text: The end-to-end suite's institution fixture server must serve only synthetic answers, over HTTPS with a certificate and key it is given, for the IBKR Flex Web Service, EODHD and NBP table A endpoints the connectors call, each below its own path that a connector base URL can name, so that the connectors' own clients read them and no end-to-end test calls a real institution.
origin:
  kind: agent
  recorded_at: '2026-11-06T22:44:09.196Z'
id: FACT-PROP-8F0DEBB9CF368C50
type: fact
---

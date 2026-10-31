---
title: Serves Broker Account Numbers = false
status: active
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-121-2
  - lane:strict
  - fact:property_value
text_ref: design-decisions:D-121#2
fact_kind: property_value
subject_key: views.read_api
property_key: serves_broker_account_numbers
operator: eq
value_type: bool
value_bool: false
canonical_key: kb-requirements-req-views-read-api-no-broker-account-numbers-md:views.read_api:serves_broker_account_numbers:eq:false
claim_key: CLAIM-1B6C42E30707A03F
claim_text: The read API must never serve broker account numbers, so that read sessions, script tokens and plugins never see one.
origin:
  kind: agent
  recorded_at: '2026-10-31T12:15:28.064Z'
id: FACT-PROP-AD6A09D5602CB79A
type: fact
---

---
title: Connectors Endpoint
status: active
text_ref: design-session:connector-endpoint-ca
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-session-connector-endpoint-ca
  - lane:strict
  - fact:subject
  - vocabulary:new-subject
fact_kind: subject
subject_key: connectors.endpoint
canonical_key: connectors.endpoint
claim_key: CLAIM-403A535044B4CC17
claim_text: Each first-party network connector, ibkr-flex, eodhd, nbp-fx and treasury-bonds for both GUS and NBP, must take an optional https base URL config field that its config schema validates and that defaults to the provider's real host, must call that base URL instead of a hard-coded host, must name the field in egressFromConfig so the grant approves a configured host, and must request the optional ca-bundle Secret its endpoint client trusts besides the system roots.
origin:
  kind: agent
  recorded_at: '2026-11-06T22:43:55.208Z'
id: FACT-SUBJECT-2989AF1C42D41A08
type: fact
---

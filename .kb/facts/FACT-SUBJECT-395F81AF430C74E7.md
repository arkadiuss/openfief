---
title: Sdk Go Connector Server
status: active
text_ref: docs/DESIGN.md#D-59
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-59
  - lane:strict
  - fact:subject
  - vocabulary:new-subject
fact_kind: subject
subject_key: sdk.go.connector_server
canonical_key: sdk.go.connector_server
claim_key: CLAIM-737BEEAD3377E06B
claim_text: The Go SDK's connector server must refuse as unauthenticated, before the connector sees it, every call without an Authorization bearer token that its core verifier accepts as the core's token for the instance, whose audience is openfief-plugin:<instance>.
origin:
  kind: agent
  recorded_at: '2026-10-28T08:01:19.098Z'
id: FACT-SUBJECT-395F81AF430C74E7
type: fact
---

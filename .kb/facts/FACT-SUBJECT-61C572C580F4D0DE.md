---
title: Sdk Endpoint
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
subject_key: sdk.endpoint
canonical_key: sdk.endpoint
claim_key: CLAIM-7C05C0F7460E7BCF
claim_text: A connector built on the Go SDK must call the base URL its config names, or the provider's real base URL when none is configured; the base URL must be an https URL of a host with an optional port and path and no user, query or fragment, plain http must be refused, and the endpoint client must refuse requests and redirects outside the base URL.
origin:
  kind: agent
  recorded_at: '2026-11-06T21:55:06.702Z'
id: FACT-SUBJECT-61C572C580F4D0DE
type: fact
---

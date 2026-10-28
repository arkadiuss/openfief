---
title: Sdk Go Egress Client
status: active
text_ref: docs/DESIGN.md#D-63
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-63
  - lane:strict
  - fact:subject
  - vocabulary:new-subject
fact_kind: subject
subject_key: sdk.go.egress_client
canonical_key: sdk.go.egress_client
claim_key: CLAIM-D2E85C5DC0970981
claim_text: The Go SDK's egress client must send every https request through the egress proxy in HTTPS_PROXY, building the proxy URL on each request with the current token from OPENFIEF_EGRESS_TOKEN_FILE (audience openfief-egress) as its password, and must refuse plain http requests, requests without a proxy, and creating a client without an egress token file.
origin:
  kind: agent
  recorded_at: '2026-10-28T08:01:53.880Z'
id: FACT-SUBJECT-5B1A524B39C882E5
type: fact
---

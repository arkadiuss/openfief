---
title: Token In Proxy Url Per Request = true
status: active
text_ref: docs/DESIGN.md#D-63
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-63
  - lane:strict
  - fact:property_value
fact_kind: property_value
subject_key: sdk.go.egress_client
property_key: token_in_proxy_url_per_request
canonical_key: mcp-kibi-model-requirement:sdk.go.egress_client:token_in_proxy_url_per_request:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-D2E85C5DC0970981
claim_text: The Go SDK's egress client must send every https request through the egress proxy in HTTPS_PROXY, building the proxy URL on each request with the current token from OPENFIEF_EGRESS_TOKEN_FILE (audience openfief-egress) as its password, and must refuse plain http requests, requests without a proxy, and creating a client without an egress token file.
origin:
  kind: agent
  recorded_at: '2026-10-28T08:01:58.341Z'
id: FACT-PROP-3893161D22A06BED
type: fact
---

---
title: Cluster Service Direct Https = true
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
subject_key: sdk.endpoint
property_key: cluster_service_direct_https
canonical_key: mcp-kibi-model-requirement:sdk.endpoint:cluster_service_direct_https:eq:true
operator: eq
value_type: bool
value_bool: true
claim_key: CLAIM-F75D54CBC2EA9CDD
claim_text: When the instance has an approved in-cluster service (D-74), the Go SDK's endpoint client must send every request directly to that service's address outside the egress proxy, over HTTPS only, verifying the service's certificate for the base URL's host against the system's roots plus the extra CA bundle, and must refuse requests outside the base URL; each first-party network connector must ask for such a service as optional.
origin:
  kind: agent
  recorded_at: '2026-11-06T23:25:47.328Z'
id: FACT-PROP-F4AF16F0046C4308
type: fact
---

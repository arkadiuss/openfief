---
title: When the instance has an approved in-cluster service (D-74), the Go SDK's endpoint client must send every request directly to that service's address outside the egress proxy, over HTTPS only, verifying the service's certificate for the base URL's host against the system's roots plus the extra CA bundle, and must refuse requests outside the base URL; each first-party network connector must ask for such a service as optional.
status: open
text_ref: design-session:connector-endpoint-ca
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-session-connector-endpoint-ca
  - lane:strict
semantic_text: When the instance has an approved in-cluster service (D-74), the Go SDK's endpoint client must send every request directly to that service's address outside the egress proxy, over HTTPS only, verifying the service's certificate for the base URL's host against the system's roots plus the extra CA bundle, and must refuse requests outside the base URL; each first-party network connector must ask for such a service as optional.
logic_claims:
  - CLAIM-F75D54CBC2EA9CDD
semantic_clauses:
  - When the instance has an approved in-cluster service (D-74), the Go SDK's endpoint client must send every request directly to that service's address outside the egress proxy, over HTTPS only, verifying the service's certificate for the base URL's host against the system's roots plus the extra CA bundle, and must refuse requests outside the base URL; each first-party network connector must ask for such a service as optional.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 9f9aca759671477442ef0622e7e75ae51793d8c013234ed8872a45e0ece99bbf
semantic_inventory:
  - claim_key: CLAIM-F75D54CBC2EA9CDD
    claim_text: When the instance has an approved in-cluster service (D-74), the Go SDK's endpoint client must send every request directly to that service's address outside the egress proxy, over HTTPS only, verifying the service's certificate for the base URL's host against the system's roots plus the extra CA bundle, and must refuse requests outside the base URL; each first-party network connector must ask for such a service as optional
    role: condition
    status: modeled
    span:
      start: 0
      end: 426
origin:
  kind: agent
  recorded_at: '2026-11-06T23:25:48.072Z'
id: REQ-sdk-endpoint-cluster-service
type: req
---
When the instance has an approved in-cluster service (D-74), the Go SDK's endpoint client must send every request directly to that service's address outside the egress proxy, over HTTPS only, verifying the service's certificate for the base URL's host against the system's roots plus the extra CA bundle, and must refuse requests outside the base URL; each first-party network connector must ask for such a service as optional.

## Context

Design session decision for M0 exit end-to-end tests (D-200) and later model providers (D-205), given to the implementation agent by the orchestrator on 2026-11-06: connectors get two config fields, a schema-validated https base URL defaulting to the real host, and an optional extra CA bundle from a Secret added to the system roots, with no test-only switches. Follow-up decision from the same design session: in-cluster fixtures have private addresses, which the egress proxy refuses (D-184), so a base URL may name either a public https host reached through the proxy with the host in the grant, or the instance approved in-cluster service reached directly (D-74). An in-cluster Ollama for D-205 uses the same path, so nothing is test-only.

## Source

> base URL may name either a public https host (egress through the proxy, host in the grant) or the instance's approved cluster service (direct, D-74).

design-session:connector-endpoint-ca

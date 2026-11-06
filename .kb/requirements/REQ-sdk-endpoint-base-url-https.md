---
title: A connector built on the Go SDK must call the base URL its config names, or the provider's real base URL when none is configured; the base URL must be an https URL of a host with an optional port and path and no user, query or fragment, plain http must be refused, and the endpoint client must refuse requests and redirects outside the base URL.
status: open
text_ref: design-session:connector-endpoint-ca
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-session-connector-endpoint-ca
  - lane:strict
semantic_text: A connector built on the Go SDK must call the base URL its config names, or the provider's real base URL when none is configured; the base URL must be an https URL of a host with an optional port and path and no user, query or fragment, plain http must be refused, and the endpoint client must refuse requests and redirects outside the base URL.
logic_claims:
  - CLAIM-7C05C0F7460E7BCF
semantic_clauses:
  - A connector built on the Go SDK must call the base URL its config names, or the provider's real base URL when none is configured; the base URL must be an https URL of a host with an optional port and path and no user, query or fragment, plain http must be refused, and the endpoint client must refuse requests and redirects outside the base URL.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: b72bcc0a6c338ec09ef0f32a12929066eb3736c2546d069444e40b080fc4e97a
semantic_inventory:
  - claim_key: CLAIM-7C05C0F7460E7BCF
    claim_text: A connector built on the Go SDK must call the base URL its config names, or the provider's real base URL when none is configured; the base URL must be an https URL of a host with an optional port and path and no user, query or fragment, plain http must be refused, and the endpoint client must refuse requests and redirects outside the base URL
    role: normative
    status: modeled
    span:
      start: 0
      end: 344
origin:
  kind: agent
  recorded_at: '2026-11-06T21:55:08.210Z'
id: REQ-sdk-endpoint-base-url-https
type: req
---
A connector built on the Go SDK must call the base URL its config names, or the provider's real base URL when none is configured; the base URL must be an https URL of a host with an optional port and path and no user, query or fragment, plain http must be refused, and the endpoint client must refuse requests and redirects outside the base URL.

## Context

Design session decision for M0 exit end-to-end tests (D-200) and later model providers (D-205), given to the implementation agent by the orchestrator on 2026-11-06: connectors get two config fields, a schema-validated https base URL defaulting to the real host, and an optional extra CA bundle from a Secret added to the system roots, with no test-only switches. The configured host reaches the grant through egressFromConfig (D-75).

## Source

> base URL: schema-validated in each connector's config JSON Schema, https only (refuse http://), defaulting to the real host; the connector calls that base URL instead of a hard-coded host.

design-session:connector-endpoint-ca

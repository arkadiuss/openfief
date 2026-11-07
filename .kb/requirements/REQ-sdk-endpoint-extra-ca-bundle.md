---
title: The Go SDK's endpoint client must trust the PEM certificates of the instance's optional ca-bundle Secret in addition to the system's root certificates, never instead of them; without the Secret it must add no certificate, and a bundle without a valid PEM certificate must be refused.
status: open
text_ref: design-session:connector-endpoint-ca
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-session-connector-endpoint-ca
  - lane:strict
semantic_text: The Go SDK's endpoint client must trust the PEM certificates of the instance's optional ca-bundle Secret in addition to the system's root certificates, never instead of them; without the Secret it must add no certificate, and a bundle without a valid PEM certificate must be refused.
logic_claims:
  - CLAIM-0414E0806AFA2613
semantic_clauses:
  - The Go SDK's endpoint client must trust the PEM certificates of the instance's optional ca-bundle Secret in addition to the system's root certificates, never instead of them; without the Secret it must add no certificate, and a bundle without a valid PEM certificate must be refused.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 22cc7b5e78d9d0c66aadf0166b388132a5a8b9805fceeafb519e8e49dba393a8
semantic_inventory:
  - claim_key: CLAIM-0414E0806AFA2613
    claim_text: The Go SDK's endpoint client must trust the PEM certificates of the instance's optional ca-bundle Secret in addition to the system's root certificates, never instead of them; without the Secret it must add no certificate, and a bundle without a valid PEM certificate must be refused
    role: normative
    status: modeled
    span:
      start: 0
      end: 282
origin:
  kind: agent
  recorded_at: '2026-11-06T21:55:10.700Z'
id: REQ-sdk-endpoint-extra-ca-bundle
type: req
---
The Go SDK's endpoint client must trust the PEM certificates of the instance's optional ca-bundle Secret in addition to the system's root certificates, never instead of them; without the Secret it must add no certificate, and a bundle without a valid PEM certificate must be refused.

## Context

Design session decision for M0 exit end-to-end tests (D-200) and later model providers (D-205), given to the implementation agent by the orchestrator on 2026-11-06: connectors get two config fields, a schema-validated https base URL defaulting to the real host, and an optional extra CA bundle from a Secret added to the system roots, with no test-only switches. Also a real feature for owners with self-hosted endpoints, such as a LAN Ollama (D-205), or a corporate TLS proxy.

## Source

> extra CA bundle from a Secret (optional): PEM bundle ADDED to the system roots, never replacing them; read by the connector's HTTP client.

design-session:connector-endpoint-ca

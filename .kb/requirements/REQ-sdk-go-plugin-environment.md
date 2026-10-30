---
title: 'The Go SDK must take a plugin''s environment from the names the operator sets: the core''s plugin port as ClusterIP:port from OPENFIEF_CORE_ADDRESS, called over plain HTTP; the token path from OPENFIEF_TOKEN_FILE, defaulting to /var/run/secrets/openfief/token; the listen port from OPENFIEF_LISTEN_PORT; and the configuration JSON from OPENFIEF_CONFIG. A missing or malformed core address or listen port must be an error.'
status: open
text_ref: docs/DESIGN.md#D-115
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-115
  - lane:strict
semantic_text: 'The Go SDK must take a plugin''s environment from the names the operator sets: the core''s plugin port as ClusterIP:port from OPENFIEF_CORE_ADDRESS, called over plain HTTP; the token path from OPENFIEF_TOKEN_FILE, defaulting to /var/run/secrets/openfief/token; the listen port from OPENFIEF_LISTEN_PORT; and the configuration JSON from OPENFIEF_CONFIG. A missing or malformed core address or listen port must be an error.'
logic_claims:
  - CLAIM-326E998054C10CA9
semantic_clauses:
  - 'The Go SDK must take a plugin''s environment from the names the operator sets: the core''s plugin port as ClusterIP:port from OPENFIEF_CORE_ADDRESS, called over plain HTTP; the token path from OPENFIEF_TOKEN_FILE, defaulting to /var/run/secrets/openfief/token; the listen port from OPENFIEF_LISTEN_PORT; and the configuration JSON from OPENFIEF_CONFIG. A missing or malformed core address or listen port must be an error.'
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: c89b17e99f98e683de362c181f719115addacad4744da9f99c126b17e54d7a3d
semantic_inventory:
  - claim_key: CLAIM-326E998054C10CA9
    claim_text: 'The Go SDK must take a plugin''s environment from the names the operator sets: the core''s plugin port as ClusterIP:port from OPENFIEF_CORE_ADDRESS, called over plain HTTP; the token path from OPENFIEF_TOKEN_FILE, defaulting to /var/run/secrets/openfief/token; the listen port from OPENFIEF_LISTEN_PORT; and the configuration JSON from OPENFIEF_CONFIG. A missing or malformed core address or listen port must be an error'
    role: normative
    status: modeled
    span:
      start: 0
      end: 418
origin:
  kind: agent
  recorded_at: '2026-10-28T08:02:35.037Z'
id: REQ-sdk-go-plugin-environment
type: req
---
The Go SDK must take a plugin's environment from the names the operator sets: the core's plugin port as ClusterIP:port from OPENFIEF_CORE_ADDRESS, called over plain HTTP; the token path from OPENFIEF_TOKEN_FILE, defaulting to /var/run/secrets/openfief/token; the listen port from OPENFIEF_LISTEN_PORT; and the configuration JSON from OPENFIEF_CONFIG. A missing or malformed core address or listen port must be an error.

## Context

D-11 gives plugins an SDK; the operator sets these variables on every plugin. D-115 makes the core address the core's plugin port, the only one plugin pods may reach, where the read API, the AuthService and the owner services are.

## Source

> The plugin port, which the Service's `connect` port targets and only plugin pods may reach, serves the read API and the services only plugins call: the AuthService (D-69) and the owner services (D-79).

docs/DESIGN.md#D-115

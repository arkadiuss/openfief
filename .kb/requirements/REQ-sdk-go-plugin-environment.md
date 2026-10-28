---
title: 'The Go SDK must take a plugin''s environment from the names the operator sets: the core''s ClusterIP:port from OPENFIEF_CORE_ADDRESS, called over plain HTTP; the token path from OPENFIEF_TOKEN_FILE, defaulting to /var/run/secrets/openfief/token; the listen port from OPENFIEF_LISTEN_PORT; and the configuration JSON from OPENFIEF_CONFIG. A missing or malformed core address or listen port must be an error.'
status: open
text_ref: docs/DESIGN.md#D-11
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-11
  - lane:strict
semantic_text: 'The Go SDK must take a plugin''s environment from the names the operator sets: the core''s ClusterIP:port from OPENFIEF_CORE_ADDRESS, called over plain HTTP; the token path from OPENFIEF_TOKEN_FILE, defaulting to /var/run/secrets/openfief/token; the listen port from OPENFIEF_LISTEN_PORT; and the configuration JSON from OPENFIEF_CONFIG. A missing or malformed core address or listen port must be an error.'
logic_claims:
  - CLAIM-B65D6BBFA0BD7609
semantic_clauses:
  - 'The Go SDK must take a plugin''s environment from the names the operator sets: the core''s ClusterIP:port from OPENFIEF_CORE_ADDRESS, called over plain HTTP; the token path from OPENFIEF_TOKEN_FILE, defaulting to /var/run/secrets/openfief/token; the listen port from OPENFIEF_LISTEN_PORT; and the configuration JSON from OPENFIEF_CONFIG. A missing or malformed core address or listen port must be an error.'
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 0a5102e287b3203e7f2f757fec469da6b69c8143f3044695f1be9c9046d40339
semantic_inventory:
  - claim_key: CLAIM-B65D6BBFA0BD7609
    claim_text: 'The Go SDK must take a plugin''s environment from the names the operator sets: the core''s ClusterIP:port from OPENFIEF_CORE_ADDRESS, called over plain HTTP; the token path from OPENFIEF_TOKEN_FILE, defaulting to /var/run/secrets/openfief/token; the listen port from OPENFIEF_LISTEN_PORT; and the configuration JSON from OPENFIEF_CONFIG. A missing or malformed core address or listen port must be an error'
    role: normative
    status: modeled
    span:
      start: 0
      end: 403
origin:
  kind: agent
  recorded_at: '2026-10-28T08:02:35.037Z'
id: REQ-sdk-go-plugin-environment
type: req
---
The Go SDK must take a plugin's environment from the names the operator sets: the core's ClusterIP:port from OPENFIEF_CORE_ADDRESS, called over plain HTTP; the token path from OPENFIEF_TOKEN_FILE, defaulting to /var/run/secrets/openfief/token; the listen port from OPENFIEF_LISTEN_PORT; and the configuration JSON from OPENFIEF_CONFIG. A missing or malformed core address or listen port must be an error.

## Context

The operator (internal/plugins/operator) sets these four variables in every plugin container when it compiles the plugin's workload (D-11). Plugins have no DNS (D-32), so the core is an address, not a name. TLS between plugins and the core is not decided, so the client uses plain HTTP inside the cluster. Supports REQ-AUTO-C6544ADAF8F5AE65. Modeled by Claude Code for the M0 plugin contracts task; the author has not reviewed it yet.

## Source

> Each plugin ships a manifest that OpenFief compiles into NetworkPolicy, Secret mounts, RBAC and egress allow-list entries (D-32).

docs/DESIGN.md D-11

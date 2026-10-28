---
title: The core must present to a plugin instance a token of its own ServiceAccount requested through the TokenRequest API with audience openfief-plugin followed by the instance name and a lifetime of 10 minutes, reuse it for that instance until 2 minutes before it expires, not cache a failed request, and never put a token in an error.
status: open
text_ref: docs/DESIGN.md#D-59
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-59
  - lane:strict
semantic_text: The core must present to a plugin instance a token of its own ServiceAccount requested through the TokenRequest API with audience openfief-plugin followed by the instance name and a lifetime of 10 minutes, reuse it for that instance until 2 minutes before it expires, not cache a failed request, and never put a token in an error.
logic_claims:
  - CLAIM-5C8E0FE7BAC6701B
semantic_clauses:
  - The core must present to a plugin instance a token of its own ServiceAccount requested through the TokenRequest API with audience openfief-plugin followed by the instance name and a lifetime of 10 minutes, reuse it for that instance until 2 minutes before it expires, not cache a failed request, and never put a token in an error.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: d20f5decd27c7119958d9476413617bddb22e4ed402d44b49200ced461395f90
semantic_inventory:
  - claim_key: CLAIM-5C8E0FE7BAC6701B
    claim_text: The core must present to a plugin instance a token of its own ServiceAccount requested through the TokenRequest API with audience openfief-plugin followed by the instance name and a lifetime of 10 minutes, reuse it for that instance until 2 minutes before it expires, not cache a failed request, and never put a token in an error
    role: normative
    status: modeled
    span:
      start: 0
      end: 329
origin:
  kind: agent
  recorded_at: '2026-10-28T17:22:48.128Z'
id: REQ-identity-core-tokens
type: req
---
The core must present to a plugin instance a token of its own ServiceAccount requested through the TokenRequest API with audience openfief-plugin followed by the instance name and a lifetime of 10 minutes, reuse it for that instance until 2 minutes before it expires, not cache a failed request, and never put a token in an error.

## Context

D-59 makes the core the only caller of plugins, proven by a short-lived TokenRequest token whose audience is the instance; the SDK expects the audience openfief.InstanceAudience. Ten minutes is the API server's minimum; the cache keeps the API server out of every call while never sending a nearly expired token. The core's Role limits TokenRequest to its own ServiceAccount. Modeled by Claude Code for the core-connectors task (orchestrator brief: CoreTokens, discovery, cmd/core wiring); the author has not reviewed this requirement yet.

## Source

> the core presents a short-lived ServiceAccount token whose audience is that plugin instance, requested through Kubernetes' TokenRequest API as instances come and go.

docs/DESIGN.md D-59

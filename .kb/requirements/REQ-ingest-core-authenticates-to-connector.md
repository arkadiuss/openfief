---
title: Every core call to a connector instance must carry, as an Authorization bearer token, a token the core's token issuer made for that instance just before the call, and the core must not call the instance when the issuer fails.
status: open
text_ref: docs/DESIGN.md#D-59
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-59
  - lane:strict
semantic_text: Every core call to a connector instance must carry, as an Authorization bearer token, a token the core's token issuer made for that instance just before the call, and the core must not call the instance when the issuer fails.
logic_claims:
  - CLAIM-4F2CFB6B6AD1A8CB
semantic_clauses:
  - Every core call to a connector instance must carry, as an Authorization bearer token, a token the core's token issuer made for that instance just before the call, and the core must not call the instance when the issuer fails.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 5af782e0700a78312b8dbc8beb84f11dd4e2b6b1b752afbe646d7057aa80ad9c
semantic_inventory:
  - claim_key: CLAIM-4F2CFB6B6AD1A8CB
    claim_text: Every core call to a connector instance must carry, as an Authorization bearer token, a token the core's token issuer made for that instance just before the call, and the core must not call the instance when the issuer fails
    role: normative
    status: modeled
    span:
      start: 0
      end: 224
origin:
  kind: agent
  recorded_at: '2026-10-28T08:56:15.385Z'
id: REQ-ingest-core-authenticates-to-connector
type: req
---
Every core call to a connector instance must carry, as an Authorization bearer token, a token the core's token issuer made for that instance just before the call, and the core must not call the instance when the issuer fails.

## Context

D-59: the core presents a short-lived ServiceAccount token whose audience is the plugin instance, requested through the TokenRequest API. The issuer is a port (app.CoreTokens) the identity context implements; the audience openfief-plugin:<instance> is defined by the Go SDK, which verifies it. Refines REQ-plugins-core-only-caller. Modeled by Claude Code for the M0 plugin contracts task; the author has not reviewed it yet.

## Source

> the core presents a short-lived ServiceAccount token whose audience is that plugin instance, requested through Kubernetes' TokenRequest API as instances come and go.

docs/DESIGN.md D-59

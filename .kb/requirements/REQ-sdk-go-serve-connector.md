---
title: The Go SDK's ServeConnector must serve a connector's contract on the plugin's listen port only to callers its core verifier accepts, refuse to start without its listen port or instance name, and on cancellation stop taking calls and let calls in flight finish within a bound.
status: open
text_ref: docs/DESIGN.md#D-59
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-59
  - lane:strict
semantic_text: The Go SDK's ServeConnector must serve a connector's contract on the plugin's listen port only to callers its core verifier accepts, refuse to start without its listen port or instance name, and on cancellation stop taking calls and let calls in flight finish within a bound.
logic_claims:
  - CLAIM-396DD329FED54C85
semantic_clauses:
  - The Go SDK's ServeConnector must serve a connector's contract on the plugin's listen port only to callers its core verifier accepts, refuse to start without its listen port or instance name, and on cancellation stop taking calls and let calls in flight finish within a bound.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 607a605d34512d9f4f8eaa3b61457ed93404436350915e8f8cb122d05b306158
semantic_inventory:
  - claim_key: CLAIM-396DD329FED54C85
    claim_text: The Go SDK's ServeConnector must serve a connector's contract on the plugin's listen port only to callers its core verifier accepts, refuse to start without its listen port or instance name, and on cancellation stop taking calls and let calls in flight finish within a bound
    role: normative
    status: modeled
    span:
      start: 0
      end: 274
origin:
  kind: agent
  recorded_at: '2026-10-28T17:07:15.242Z'
id: REQ-sdk-go-serve-connector
type: req
---
The Go SDK's ServeConnector must serve a connector's contract on the plugin's listen port only to callers its core verifier accepts, refuse to start without its listen port or instance name, and on cancellation stop taking calls and let calls in flight finish within a bound.

## Context

D-49 and D-59: a connector serves a Connect service that only the core may call. Putting the whole serving side in the SDK keeps first-party and third-party plugin binaries thin and makes the core-only check impossible to forget. A bounded graceful shutdown lets a pod roll during an import batch without cutting a call mid-way; the core retries failed batches anyway. Refines REQ-plugins-core-only-caller. Modeled by Claude Code for the M0 plugin binaries task (agent A: SDK verifier, operator key publishing, first-party plugins); the author has not reviewed it yet.

## Source

> Only the core may call a plugin. A plugin's NetworkPolicy admits ingress from the core alone, and the core presents a short-lived ServiceAccount token whose audience is that plugin instance, requested through Kubernetes' TokenRequest API as instances come and go. The SDK verifies it. Plugins can't reach the Kubernetes API, so the operator publishes the cluster's token-signing public keys in a ConfigMap the SDK reads.

docs/DESIGN.md D-59

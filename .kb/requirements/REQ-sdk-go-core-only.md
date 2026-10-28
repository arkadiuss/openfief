---
title: The Go SDK's connector server must refuse as unauthenticated, before the connector sees it, every call without an Authorization bearer token that its core verifier accepts as the core's token for the instance, whose audience is openfief-plugin:<instance>.
status: open
text_ref: docs/DESIGN.md#D-59
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-59
  - lane:strict
semantic_text: The Go SDK's connector server must refuse as unauthenticated, before the connector sees it, every call without an Authorization bearer token that its core verifier accepts as the core's token for the instance, whose audience is openfief-plugin:<instance>.
logic_claims:
  - CLAIM-737BEEAD3377E06B
semantic_clauses:
  - The Go SDK's connector server must refuse as unauthenticated, before the connector sees it, every call without an Authorization bearer token that its core verifier accepts as the core's token for the instance, whose audience is openfief-plugin:<instance>.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 34bfe0228c6cfa21ad797ad77226a4ef3d3e6e31aadba7a782f929036597e1cd
semantic_inventory:
  - claim_key: CLAIM-737BEEAD3377E06B
    claim_text: The Go SDK's connector server must refuse as unauthenticated, before the connector sees it, every call without an Authorization bearer token that its core verifier accepts as the core's token for the instance, whose audience is openfief-plugin:<instance>
    role: normative
    status: modeled
    span:
      start: 0
      end: 254
origin:
  kind: agent
  recorded_at: '2026-10-28T08:01:30.809Z'
id: REQ-sdk-go-core-only
type: req
---
The Go SDK's connector server must refuse as unauthenticated, before the connector sees it, every call without an Authorization bearer token that its core verifier accepts as the core's token for the instance, whose audience is openfief-plugin:<instance>.

## Context

D-59: only the core may call a plugin, presenting a short-lived ServiceAccount token whose audience is that plugin instance, and the SDK verifies it. The audience string openfief-plugin:<instance> is the SDK's choice; how the token is verified (TokenReview in M0 per the author's orchestrator, or published signing keys as D-59 describes) sits behind the CoreVerifier interface. Refines REQ-plugins-core-only-caller. Modeled by Claude Code for the M0 plugin contracts task; the author has not reviewed it yet.

## Source

> Only the core may call a plugin. [...] the core presents a short-lived ServiceAccount token whose audience is that plugin instance, requested through Kubernetes' TokenRequest API as instances come and go. The SDK verifies it.

docs/DESIGN.md D-59

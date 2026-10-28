---
title: 'A plugin must accept calls only from the core: its NetworkPolicy must admit ingress from the core alone, and each call must carry a short-lived ServiceAccount token whose audience is that plugin instance.'
status: open
text_ref: design-decisions:D-59#1
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:design-decisions-d-59-1
  - lane:strict
semantic_text: 'A plugin must accept calls only from the core: its NetworkPolicy must admit ingress from the core alone, and each call must carry a short-lived ServiceAccount token whose audience is that plugin instance.'
logic_claims:
  - CLAIM-CF03A08EF5A303C7
semantic_clauses:
  - 'A plugin must accept calls only from the core: its NetworkPolicy must admit ingress from the core alone, and each call must carry a short-lived ServiceAccount token whose audience is that plugin instance.'
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 6b14d26b53b9a63f4e2d0cc6863e9b7ed11e39e7f2046ea100a8414d5f1ec99c
semantic_inventory:
  - claim_key: CLAIM-CF03A08EF5A303C7
    claim_text: 'A plugin must accept calls only from the core: its NetworkPolicy must admit ingress from the core alone, and each call must carry a short-lived ServiceAccount token whose audience is that plugin instance'
    role: normative
    status: modeled
    span:
      start: 0
      end: 203
firmness: firm
required_evidence: asserted
origin:
  kind: import
  ref: design-decisions:D-59#1
  recorded_at: '2026-10-28T06:09:40.027Z'
id: REQ-plugins-core-only-caller
type: req
---
A plugin must accept calls only from the core: its NetworkPolicy must admit ingress from the core alone, and each call must carry a short-lived ServiceAccount token whose audience is that plugin instance.

## Context

GR-12 puts the security model first, per use case. M0 has one use case, tracking and imports; the use cases of later milestones are in the proposal files.

## Source

> Only the core may call a plugin. A plugin's NetworkPolicy admits ingress from the core alone, and the core presents a short-lived ServiceAccount token whose audience is that plugin instance, requested through Kubernetes' TokenRequest API as instances come and go. The SDK verifies it. Plugins can't reach the Kubernetes API, so the operator publishes the cluster's token-signing public keys in a ConfigMap the SDK reads.

OpenFief author decisions (docs/DESIGN.md, section 3.7), design-decisions:D-59#1

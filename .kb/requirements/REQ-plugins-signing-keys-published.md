---
title: The operator must publish the cluster's ServiceAccount token issuer and only public RSA or EC signing keys with key IDs, read from the API server's issuer discovery endpoints, in the ConfigMap openfief-sa-jwks at start, periodically and whenever the ConfigMap changes, keeping the published keys when they can't be read.
status: open
text_ref: docs/DESIGN.md#D-59
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-59
  - lane:strict
semantic_text: The operator must publish the cluster's ServiceAccount token issuer and only public RSA or EC signing keys with key IDs, read from the API server's issuer discovery endpoints, in the ConfigMap openfief-sa-jwks at start, periodically and whenever the ConfigMap changes, keeping the published keys when they can't be read.
logic_claims:
  - CLAIM-BDD7C3912DC1D6A1
semantic_clauses:
  - The operator must publish the cluster's ServiceAccount token issuer and only public RSA or EC signing keys with key IDs, read from the API server's issuer discovery endpoints, in the ConfigMap openfief-sa-jwks at start, periodically and whenever the ConfigMap changes, keeping the published keys when they can't be read.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: a93e024daab47830bb710fbc6dcf6d16ce32d12073a2c91eced7a40c32d5eae9
semantic_inventory:
  - claim_key: CLAIM-BDD7C3912DC1D6A1
    claim_text: The operator must publish the cluster's ServiceAccount token issuer and only public RSA or EC signing keys with key IDs, read from the API server's issuer discovery endpoints, in the ConfigMap openfief-sa-jwks at start, periodically and whenever the ConfigMap changes, keeping the published keys when they can't be read
    role: normative
    status: modeled
    span:
      start: 0
      end: 319
origin:
  kind: agent
  recorded_at: '2026-10-28T16:47:42.766Z'
id: REQ-plugins-signing-keys-published
type: req
---
The operator must publish the cluster's ServiceAccount token issuer and only public RSA or EC signing keys with key IDs, read from the API server's issuer discovery endpoints, in the ConfigMap openfief-sa-jwks at start, periodically and whenever the ConfigMap changes, keeping the published keys when they can't be read.

## Context

D-59: plugins can't reach the Kubernetes API, so the operator publishes the token-signing public keys in a ConfigMap the SDK reads. The keys come from /openid/v1/jwks and the issuer from /.well-known/openid-configuration, non-resource URLs that Kubernetes' default ClusterRoleBinding system:service-account-issuer-discovery grants to every ServiceAccount, so the namespace-scoped operator (D-15) needs no ClusterRole. Refreshing periodically follows key rotation; re-applying on change restores an edited or deleted ConfigMap. Refusing private or symmetric material keeps a misbehaving source from handing plugins a key they could be fooled with. Refines REQ-plugins-core-only-caller. Modeled by Claude Code for the M0 plugin binaries task (agent A: SDK verifier, operator key publishing, first-party plugins); the author has not reviewed it yet.

## Source

> Only the core may call a plugin. A plugin's NetworkPolicy admits ingress from the core alone, and the core presents a short-lived ServiceAccount token whose audience is that plugin instance, requested through Kubernetes' TokenRequest API as instances come and go. The SDK verifies it. Plugins can't reach the Kubernetes API, so the operator publishes the cluster's token-signing public keys in a ConfigMap the SDK reads.

docs/DESIGN.md D-59

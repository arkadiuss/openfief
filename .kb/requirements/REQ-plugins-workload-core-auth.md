---
title: Every plugin's workload must mount the ConfigMap openfief-sa-jwks read-only and set OPENFIEF_INSTANCE to the instance's name, OPENFIEF_CORE_SUBJECT to the core's ServiceAccount username and OPENFIEF_SA_JWKS_DIR to the mount, so the plugin can verify the core's token without Kubernetes API access.
status: open
text_ref: docs/DESIGN.md#D-59
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-59
  - lane:strict
semantic_text: Every plugin's workload must mount the ConfigMap openfief-sa-jwks read-only and set OPENFIEF_INSTANCE to the instance's name, OPENFIEF_CORE_SUBJECT to the core's ServiceAccount username and OPENFIEF_SA_JWKS_DIR to the mount, so the plugin can verify the core's token without Kubernetes API access.
logic_claims:
  - CLAIM-A799BE6B51BB4F8F
semantic_clauses:
  - Every plugin's workload must mount the ConfigMap openfief-sa-jwks read-only and set OPENFIEF_INSTANCE to the instance's name, OPENFIEF_CORE_SUBJECT to the core's ServiceAccount username and OPENFIEF_SA_JWKS_DIR to the mount, so the plugin can verify the core's token without Kubernetes API access.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: 93a87c20d1f0c8b97ec90d2f4ae18921381dc24b6edade02e5d2dea78c0619fc
semantic_inventory:
  - claim_key: CLAIM-A799BE6B51BB4F8F
    claim_text: Every plugin's workload must mount the ConfigMap openfief-sa-jwks read-only and set OPENFIEF_INSTANCE to the instance's name, OPENFIEF_CORE_SUBJECT to the core's ServiceAccount username and OPENFIEF_SA_JWKS_DIR to the mount, so the plugin can verify the core's token without Kubernetes API access
    role: normative
    status: modeled
    span:
      start: 0
      end: 296
origin:
  kind: agent
  recorded_at: '2026-10-28T16:47:47.686Z'
id: REQ-plugins-workload-core-auth
type: req
---
Every plugin's workload must mount the ConfigMap openfief-sa-jwks read-only and set OPENFIEF_INSTANCE to the instance's name, OPENFIEF_CORE_SUBJECT to the core's ServiceAccount username and OPENFIEF_SA_JWKS_DIR to the mount, so the plugin can verify the core's token without Kubernetes API access.

## Context

D-59: the SDK verifies the core's token, whose audience is the plugin instance (D-48), from keys the operator publishes. The plugin therefore needs its own instance name, the core's identity and the keys, and it has no API access to look any of them up. The mount is not optional, so a plugin pod never starts without keys and refuses every call instead. The core's ServiceAccount is an operator setting, since installations may name it. Refines REQ-plugins-core-only-caller and REQ-plugins-instance-per-install. Modeled by Claude Code for the M0 plugin binaries task (agent A: SDK verifier, operator key publishing, first-party plugins); the author has not reviewed it yet.

## Source

> Only the core may call a plugin. A plugin's NetworkPolicy admits ingress from the core alone, and the core presents a short-lived ServiceAccount token whose audience is that plugin instance, requested through Kubernetes' TokenRequest API as instances come and go. The SDK verifies it. Plugins can't reach the Kubernetes API, so the operator publishes the cluster's token-signing public keys in a ConfigMap the SDK reads.

docs/DESIGN.md D-59

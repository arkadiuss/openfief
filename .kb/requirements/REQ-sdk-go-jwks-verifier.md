---
title: The Go SDK's core verifier must accept a bearer token only when it is signed with RS256 or ES256, ES384 or ES512 by a key in the mounted cluster key set, names the mounted issuer, has the core's ServiceAccount as subject and the instance's audience, and is unexpired; alg none and HMAC must be refused.
status: open
text_ref: docs/DESIGN.md#D-59
tags:
  - strict-modeling
  - confidence:0.90
  - confidence-band:high
  - provenance:docs-design-md-d-59
  - lane:strict
semantic_text: The Go SDK's core verifier must accept a bearer token only when it is signed with RS256 or ES256, ES384 or ES512 by a key in the mounted cluster key set, names the mounted issuer, has the core's ServiceAccount as subject and the instance's audience, and is unexpired; alg none and HMAC must be refused.
logic_claims:
  - CLAIM-D1AEC26052D663F0
semantic_clauses:
  - The Go SDK's core verifier must accept a bearer token only when it is signed with RS256 or ES256, ES384 or ES512 by a key in the mounted cluster key set, names the mounted issuer, has the core's ServiceAccount as subject and the instance's audience, and is unexpired; alg none and HMAC must be refused.
semantic_inventory_version: kibi.semantic-inventory.v1
semantic_source_field: semantic_text
semantic_source_hash: b796c8bf8c8c0efde3efcce9db16b38c34162e10ace18be87f41e17dfcf1305b
semantic_inventory:
  - claim_key: CLAIM-D1AEC26052D663F0
    claim_text: The Go SDK's core verifier must accept a bearer token only when it is signed with RS256 or ES256, ES384 or ES512 by a key in the mounted cluster key set, names the mounted issuer, has the core's ServiceAccount as subject and the instance's audience, and is unexpired; alg none and HMAC must be refused
    role: normative
    status: modeled
    span:
      start: 0
      end: 301
origin:
  kind: agent
  recorded_at: '2026-10-28T16:47:35.858Z'
id: REQ-sdk-go-jwks-verifier
type: req
---
The Go SDK's core verifier must accept a bearer token only when it is signed with RS256 or ES256, ES384 or ES512 by a key in the mounted cluster key set, names the mounted issuer, has the core's ServiceAccount as subject and the instance's audience, and is unexpired; alg none and HMAC must be refused.

## Context

D-59 has the SDK verify the core's per-instance token with the cluster's token-signing public keys that the operator publishes, because plugins can't reach the Kubernetes API for a TokenReview. The token's signature proves the cluster issued it, the subject proves it is the core's ServiceAccount, and the audience proves it was requested for this instance, so one plugin can't replay a token meant for another. Only the algorithms Kubernetes signs with are allowed, which refuses alg none and HMAC key confusion before any key is tried (D-29: a maintained JOSE library, go-jose v4, no home-grown cryptography). The key files are re-read on every call and parsed again only when they change, so rotated keys apply as soon as the kubelet updates the mount. Refines REQ-plugins-core-only-caller. Modeled by Claude Code for the M0 plugin binaries task (agent A: SDK verifier, operator key publishing, first-party plugins); the author has not reviewed it yet.

## Source

> Only the core may call a plugin. A plugin's NetworkPolicy admits ingress from the core alone, and the core presents a short-lived ServiceAccount token whose audience is that plugin instance, requested through Kubernetes' TokenRequest API as instances come and go. The SDK verifies it. Plugins can't reach the Kubernetes API, so the operator publishes the cluster's token-signing public keys in a ConfigMap the SDK reads.

docs/DESIGN.md D-59

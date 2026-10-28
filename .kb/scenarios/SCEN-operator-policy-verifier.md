---
title: Plugins run only with a trusted signature and matching manifest
status: active
origin:
  kind: agent
  recorded_at: '2026-10-28T09:29:24.969Z'
id: SCEN-operator-policy-verifier
type: scenario
---
Given a Plugin and its grant, when its image is signed by OpenFief's CI and its signed manifest label equals its manifest, then it runs and its SignatureVerified condition is True with reason SignedByTrustedSigner; when the label differs, then it does not run and the condition is False with reason ManifestMismatch; when the registry cannot be read, the condition is Unknown and the check is retried.
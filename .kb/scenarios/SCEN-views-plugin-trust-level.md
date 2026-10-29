---
title: Each installed plugin shows one trust level derived from its status
status: active
origin:
  kind: agent
  recorded_at: '2026-10-29T03:29:28.083Z'
id: SCEN-views-plugin-trust-level
type: scenario
---
Given a Plugin's SignatureVerified condition and status.signature, when the core maps them, then a True condition is verified, signed by OpenFief's CI, by a catalog publisher, by the owner, or by a trusted signer when the kind is missing; False is refused; Unknown or no condition is not checked; and the signer's name is shown only for a trusted signature.
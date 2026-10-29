---
title: A trusted signature names its signer and kind in the Plugin's status
status: active
origin:
  kind: agent
  recorded_at: '2026-10-29T03:01:14.977Z'
id: SCEN-operator-signature-signer-status
type: scenario
---
Given a Plugin whose image a trusted signer signed, when the operator verifies it, then the SignatureVerified condition is True and status.signature holds Verified for a catalog-pinned image signed by its publisher, or Unverified with the signer kind OpenFiefCI, Publisher or Owner and the signer's name; a refused or unchecked image has no status.signature.
---
title: The operator verifies a key-signed plugin image and reports it
status: active
origin:
  kind: agent
  recorded_at: '2026-10-29T10:09:40.524Z'
id: SCEN-e2e-plugin-signature-verified
type: scenario
---
Given the nbp-fx image in the suite's registry signed with the run's key, which the owner's trusted signers list, when a Plugin and its PluginGrant are applied, then the Plugin becomes Ready with SignatureVerified True for reason SignedByTrustedSigner and status.signature naming the owner's signer with trust Unverified.